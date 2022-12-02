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
	cdn_errors "animakuro/cdn/internal/cdn/errors"
	cdnpath "animakuro/cdn/internal/cdn/path"
	cdnutil "animakuro/cdn/internal/cdn/util"
	"animakuro/cdn/internal/cdn/validate"
	"animakuro/cdn/internal/entities"
	"animakuro/cdn/internal/formdata"
	"animakuro/cdn/internal/fs"
	"animakuro/cdn/internal/modules"
	bucketcache "animakuro/cdn/pkg/cache/bucket"
	filecache "animakuro/cdn/pkg/cache/file"
	"animakuro/cdn/pkg/hash"
	"animakuro/cdn/pkg/middleware"

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
	bc          *bucketcache.BucketCache
	fc          filecache.Incrementer
}

func NewHandler(logger *zap.SugaredLogger,
	mux *mux.Router,
	service Service,
	memConfig *config.MemoryConfig,
	middlewares *middleware.Middlewares,
	bucketCache *bucketcache.BucketCache,
	fileCache filecache.Incrementer) *Handler {
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

	// Check if bucket is registered
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

	// TODO: move to service
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

	// Add to bucket cache
	h.bc.Add(b)

	response.Created(w)
	return
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	bucket := vars[cdn_go.BucketKey]
	uuid := vars[cdn_go.FileUUIDKey]

	var rawQuery string
	var isOriginal bool

	// Get bucket from cache
	b, err := h.bc.Get(bucket)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	// Convert URL to moduleMap (see modules.Parse impl)
	moduleMap, err := modules.Parse(r.URL.Query(), b.Module)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	isOriginal = moduleMap == nil
	rawQuery = modules.Raw(moduleMap, uuid)

	sha1 := hash.SHA1Name(rawQuery)

	if !isOriginal {
		// Make path to file and check if already resolved file exists. (isOriginal = false)
		pathToExisting := cdnpath.ToExistingFile(&cdnpath.Existing{
			BucketsPath: fs.BucketsPath(),
			Bucket:      bucket,
			UUID:        uuid,
			SHA1:        sha1,
		})

		bits, isAvailable, err := h.service.TryReadExisting(pathToExisting)
		if err != nil {
			cdn_errors.ToHttp(h.logger, w, err)
			return
		}

		// Available locally
		if isAvailable {
			h.fc.Increment(pathToExisting)
			response.Binary(w, bits, h.service.ParseMime(bits))
			return
		}
	}

	// Get original file meta from DB
	f, err := h.service.GetFileDB(r.Context(), bucket, uuid)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)

		// If meta is not found in DB - delete file from disk.
		if errors.Is(err, entities.ErrFileNotFound) {
			dirPath := cdnpath.ToDir(bucket, uuid)
			h.service.TryDeleteLocally(dirPath)
		}

		return
	}

	// Make path to original file in disk
	pathToOriginal := cdnpath.ToOriginalFile(&cdnpath.Original{
		BucketsPath: fs.BucketsPath(),
		Bucket:      bucket,
		UUID:        uuid,
		DefaultName: fs.DefaultName + f.Extension,
	})
	bits, err := h.service.ReadFile(isOriginal, pathToOriginal, f.AvailableIn)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	// Serve original file without module processing (isOriginal = true)
	if isOriginal {
		//Can use original file's (f) MimeType
		response.Binary(w, bits, f.MimeType)
		h.fc.Increment(pathToOriginal)
		h.logger.Debugf("serving original file: %s", pathToOriginal)
		return
	}

	// Magic happens here
	// UseResolver would modify buff according to moduleMap
	buff := bytes.NewBuffer(bits)
	err = modules.UseResolvers(buff, b.Module, moduleMap)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	// Make path to resolved file in disk after service.MustSave
	pathToResolved := path.Join(fs.BucketsPath(), f.Bucket, uuid, sha1)
	h.fc.Increment(pathToResolved)

	buffBits := buff.Bytes()

	defer h.service.MustSave(buffBits, pathToResolved)

	response.Binary(w, buffBits, h.service.ParseMime(buffBits))
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	bucket := vars[cdn_go.BucketKey]

	err := r.ParseMultipartForm(h.memConfig.MaxUploadSize)
	if err != nil {
		err = cdnutil.WrapInternal(err, "Handler.Upload.r.ParseMultipartForm")
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	// Get files attached to MultipartForm
	files, err := formdata.ParseFiles(r.MultipartForm)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	// Upload files to bucket
	urls, ids, err := h.service.UploadMany(r.Context(), bucket, files)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	response.Json(h.logger, w, http.StatusCreated, response.JSON{
		"ids":  ids,
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

	// Delete file meta in DB
	err = h.service.DeleteFileDB(r.Context(), bucket, uuid)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	// Make path to dir containing all resolved and original files in disk
	dirPath := path.Join(fs.BucketsPath(), bucket, uuid)
	err = h.service.DeleteAll(dirPath)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	//TODO: maybe clear from cache here...

	response.Ok(w)
}