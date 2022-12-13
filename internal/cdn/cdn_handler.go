package cdn

import (
	"bytes"
	"encoding/json"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"path"

	cdn_go "animakuro/cdn"
	"animakuro/cdn/config"
	"animakuro/cdn/internal/cdn/cdnutil"
	"animakuro/cdn/internal/cdn/dto"
	cdn_errors "animakuro/cdn/internal/cdn/errors"
	cdnpath "animakuro/cdn/internal/cdn/path"
	"animakuro/cdn/internal/cdn/validate"
	"animakuro/cdn/internal/entities"
	"animakuro/cdn/internal/formdata"
	"animakuro/cdn/internal/fs"
	"animakuro/cdn/internal/modules"
	bucketcache "animakuro/cdn/pkg/cache/bucket"
	filecache "animakuro/cdn/pkg/cache/file"
	"animakuro/cdn/pkg/hash"
	"animakuro/cdn/pkg/http/response"
	"animakuro/cdn/pkg/middleware"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Handler struct {
	logger           *zap.SugaredLogger
	service          Service
	mux              *mux.Router
	memConfig        *config.MemoryConfig
	middlewares      *middleware.Middlewares
	moduleController modules.Controller
	bc               *bucketcache.BucketCache
	fc               filecache.FileCache
}

type HandlerDeps struct {
	Logger           *zap.SugaredLogger
	Mux              *mux.Router
	Middlewares      *middleware.Middlewares
	Service          Service
	ModuleController modules.Controller
	BucketCache      *bucketcache.BucketCache
	FileCache        filecache.FileCache
	MemConfig        *config.MemoryConfig
}

func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{
		logger:           deps.Logger,
		mux:              deps.Mux,
		service:          deps.Service,
		memConfig:        deps.MemConfig,
		middlewares:      deps.Middlewares,
		moduleController: deps.ModuleController,
		bc:               deps.BucketCache,
		fc:               deps.FileCache,
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
	if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
		err = cdnutil.WrapInternal(err, "Handler.CreateBucket.json.Decode")
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	// Validations
	{
		if err := validate.ValidateRequiredFields(inp); err != nil {
			cdn_errors.ToHttp(h.logger, w, err)
			return
		}

		if err := validate.BucketOperation(inp.Operations); err != nil {
			cdn_errors.ToHttp(h.logger, w, err)
			return
		}

		if ok := h.moduleController.DoesModuleExist(inp.Module); !ok {
			cdn_errors.ToHttp(h.logger, w, modules.ErrNotFound)
			return
		}
	}

	// Also checks if exists locally
	if err := fs.CreateBucket(inp.Name); err != nil {
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
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	bucket := vars[cdn_go.BucketKey]
	uuid := vars[cdn_go.FileUUIDKey]

	var (
		rawQuery   string
		isOriginal bool
	)

	// Get bucket from cache
	b, err := h.bc.Get(bucket)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	// Convert URL to moduleMap (see modules.Parse impl)
	moduleMap, err := h.moduleController.Parse(r.URL.Query(), b.Module)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	isOriginal = moduleMap == nil
	rawQuery = h.moduleController.Raw(moduleMap, uuid)
	sha1 := hash.SHA1Name(rawQuery)

	// Try go get exiting processed file.
	if !isOriginal {
		// Make path to file and check if already resolved file exists. (isOriginal = false)
		pathToExisting := cdnpath.ToExistingFile(&cdnpath.Existing{
			BucketsPath: fs.BucketsPath(),
			Bucket:      bucket,
			UUID:        uuid,
			SHA1:        sha1,
		})

		bits, isAvailable, err := h.service.ReadExisting(pathToExisting)
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
			// TODO: mark for deletion
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

	bits, err := h.service.ReadFile(pathToOriginal, f.AvailableIn)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	// Serve original file without module processing (isOriginal = true)
	if isOriginal {
		// Can use original file's (f) MimeType
		response.Binary(w, bits, f.MimeType)
		h.fc.Increment(pathToOriginal)
		h.logger.Debugf("serving original file: %s", pathToOriginal)
		return
	}

	// Magic happens here
	// UseResolver would modify buff according to moduleMap
	// TODO: think for resolving queue
	buff := bytes.NewBuffer(bits)
	err = h.moduleController.UseResolvers(buff, b.Module, moduleMap)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	// Make path to resolved file in disk after service.MustSave
	pathToResolved := path.Join(fs.BucketsPath(), f.Bucket, uuid, sha1)
	h.fc.Increment(pathToResolved)

	buffBits := buff.Bytes()
	// Important to execute asynchronously (defer), due to
	// if it fails somehow the next call to Get with resolvers will
	// resolve (process) the file again and try to save once more.
	// There's no need to save synchronously. Client will get it's file bits no matter what.
	response.Binary(w, buffBits, h.service.ParseMime(buffBits))
	h.service.MustSave(buffBits, pathToResolved)
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

	// todo: iterate over f.AvailableIn and send the same delete request...
	_ = f

	// Already marked
	if f.IsDeletable == true {
		cdn_errors.ToHttp(h.logger, w, entities.ErrFileAlreadyDeleted)
		return
	}

	// Mark file as deletable in DB
	err = h.service.MarkAsDeletableDB(r.Context(), bucket, f.ID)
	if err != nil {
		cdn_errors.ToHttp(h.logger, w, err)
		return
	}

	//TODO: maybe clear from cache here...
	response.Ok(w)
}
