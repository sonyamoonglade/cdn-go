package cdn

import (
	"bytes"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"path"

	cdn_go "animakuro/cdn"
	"animakuro/cdn/config"
	"animakuro/cdn/internal/cdn/dto"
	"animakuro/cdn/internal/formdata"
	"animakuro/cdn/internal/fs"
	"animakuro/cdn/internal/modules"
	cache "animakuro/cdn/pkg/cache/bucket"
	filecache "animakuro/cdn/pkg/cache/file"
	"animakuro/cdn/pkg/cdn_errors"
	"animakuro/cdn/pkg/hash"
	"animakuro/cdn/pkg/middleware"
	"animakuro/cdn/pkg/validate"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sonyamoonglade/delivery-service/pkg/binder"
	"github.com/sonyamoonglade/notification-service/pkg/response"
	"go.uber.org/zap"
)

type Handler struct {
	logger      *zap.SugaredLogger
	service     Service
	mux         *mux.Router
	memConfig   *config.MemoryConfig
	middlewares *middleware.Middlewares
	bc          *cache.BucketCache
	fc          *filecache.FileCache
}

func NewHandler(logger *zap.SugaredLogger,
	mux *mux.Router,
	service Service,
	memConfig *config.MemoryConfig,
	middlewares *middleware.Middlewares,
	bucketCache *cache.BucketCache,
	fileCache *filecache.FileCache) *Handler {
	return &Handler{
		logger:      logger,
		mux:         mux,
		service:     service,
		memConfig:   memConfig,
		middlewares: middlewares,
		bc:          bucketCache,
		fc:          fileCache,
	}
}

func (h *Handler) InitRoutes() {

	//shorthands for middlewares
	auth := h.middlewares.JwtMiddleware.Auth

	api := h.mux.PathPrefix("/api").Subrouter()
	{
		api.HandleFunc("/health", h.Healthcheck).Methods(http.MethodGet)
		//todo: get rid of it
		api.HandleFunc("/bucket", h.CreateBucket).Methods(http.MethodPost)
	}

	//cdn routes
	h.mux.HandleFunc("/{bucket}", auth(h.Upload)).Methods(http.MethodPost)
	h.mux.HandleFunc("/{bucket}/{fileUUID}", auth(h.Get)).Methods(http.MethodGet)
	h.mux.HandleFunc("/{bucket}/{fileUUID}", auth(h.Delete)).Methods(http.MethodDelete)
}

func (h *Handler) Healthcheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	return
}
func (h *Handler) CreateBucket(w http.ResponseWriter, r *http.Request) {

	var inp dto.CreateBucketDto
	if err := binder.Bind(r.Body, &inp); err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	err := modules.DoesModuleExist(inp.Module)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	err = validate.BucketOperation(inp.Operations)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	err = fs.CreateBucket(inp.Name)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	b, err := h.service.SaveBucketDB(r.Context(), inp)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	response.Created(w)
	h.bc.Add(b)
	return
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	bucket := vars[cdn_go.BucketKey]
	uuid := vars[cdn_go.FileUUIDKey]

	var rawQuery string
	var isOriginal bool

	b, err := h.bc.Get(bucket)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	moduleMap, err := modules.Parse(r.URL.Query(), b.Module)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	isOriginal = moduleMap == nil
	rawQuery = modules.Raw(moduleMap, uuid)

	sha1 := hash.SHA1Name(rawQuery)

	if !isOriginal {
		pathToExisting := path.Join(fs.BucketsPath(), bucket, uuid, sha1)
		bits, isAvailable, err := h.service.TryReadExisting(pathToExisting)

		if err != nil {
			cdn_errors.ToHttp(h.logger, w, err)
		}

		if isAvailable {
			h.fc.Hit(pathToExisting)
			response.Binary(w, bits, h.service.ParseMime(bits))
			return
		}
	}

	//Get original file
	f, err := h.service.GetFileDB(r.Context(), bucket, uuid)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		if errors.Is(err, cdn_errors.ErrFileNotFound) {
			//Try to delete file locally (may exist)
			dirPath := path.Join(bucket, uuid)
			h.service.TryDeleteLocally(dirPath)
		}
		return
	}

	pathToOriginal := path.Join(fs.BucketsPath(), bucket, uuid, fs.DefaultName+f.Extension)
	bits, err := h.service.ReadFile(isOriginal, pathToOriginal, f.AvailableIn)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	//Serve original file without module processing
	if isOriginal {
		//Can use original file's (f) MimeType
		response.Binary(w, bits, f.MimeType)
		h.fc.Hit(pathToOriginal)
		h.logger.Debugf("serving original file: %s", pathToOriginal)
		return
	}

	//todo: mime processing
	buff := bytes.NewBuffer(bits)
	err = modules.UseResolvers(buff, b.Module, moduleMap)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	//Represents path (in future) of resolved file
	pathToResolved := path.Join(fs.BucketsPath(), f.Bucket, uuid, sha1)
	h.fc.Hit(pathToResolved)

	buffBits := buff.Bytes()
	response.Binary(w, buffBits, h.service.ParseMime(buffBits))

	defer h.service.MustSave(buff.Bytes(), pathToResolved)
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	bucket := vars[cdn_go.BucketKey]

	err := r.ParseMultipartForm(h.memConfig.MaxUploadSize)
	if err != nil {
		err = cdn_errors.WrapInternal(err, "Handler.Upload.r.ParseMultipartForm")
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	files, err := formdata.ParseFiles(r.MultipartForm)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	urls, err := h.service.UploadMany(r.Context(), bucket, files)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	response.Json(h.logger, w, http.StatusCreated, response.JSON{
		"urls": urls,
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	bucket := vars[cdn_go.BucketKey]
	uuid := vars[cdn_go.FileUUIDKey]

	f, err := h.service.GetFileDB(r.Context(), bucket, uuid)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	//todo: iterate over f.AvailableIn and send the same delete request...
	_ = f

	err = h.service.DeleteFileDB(r.Context(), bucket, uuid)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	dirPath := path.Join(fs.BucketsPath(), bucket, uuid)
	err = h.service.DeleteAll(dirPath)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	response.Ok(w)
}
