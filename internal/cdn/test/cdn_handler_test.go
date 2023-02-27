package cdn

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"animakuro/cdn/internal/cdn"
	mock_cdn "animakuro/cdn/internal/cdn/mocks"
	"animakuro/cdn/internal/entities"
	"animakuro/cdn/internal/modules"
	mock_modules "animakuro/cdn/internal/modules/mocks"
	bucketcache "animakuro/cdn/pkg/cache/bucket"
	filecache "animakuro/cdn/pkg/cache/file"

	"github.com/gabriel-vasile/mimetype"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

var bucket = &entities.Bucket{
	ID:   primitive.ObjectID{},
	Name: "site-content",
	Operations: []*entities.Operation{
		{
			Name: "get",
			Type: "private",
			Keys: []string{"abcd"},
		},
		{
			Name: "post",
			Type: "private",
			Keys: []string{"abcd"},
		},
	},
	Module: "image",
}

// Do not use t.Parallel(). It breaks mocking with EXPECT()
func TestGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service, moduleController := getMocks(ctrl)
	deps := setupDeps()

	router := deps.Mux

	handler := cdn.NewHandler(&cdn.HandlerDeps{
		Logger:           deps.Logger,
		Mux:              deps.Mux,
		BucketCache:      deps.BucketCache,
		FileCache:        deps.FileCache,
		Service:          service,
		ModuleController: moduleController,
		// Pass nil: see cdn_handler_test.go:97
		Middlewares: nil,
		MemConfig:   nil,
	})

	router.HandleFunc("/{bucket}/{fileUUID}", handler.Get)

	// build path to resource
	fileID := uuid.NewString()

	t.Run("should return error. Requested bucket does not exist", func(t *testing.T) {
		url := fmt.Sprintf("https://cdn.com/%s/%s", "abracadabra" /* bucket */, fileID /* uuid */)
		r, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()

		// Will call handler.Get
		router.ServeHTTP(w, r)

		respStr := w.Body.String()
		code := w.Result().StatusCode

		expectedResponse := fmt.Sprintf(`{"message":"%s"}`, entities.ErrBucketNotFound)
		require.Equal(t, expectedResponse, respStr)
		require.Equal(t, http.StatusNotFound, code)
	})

	t.Run("should get original file that exists", func(t *testing.T) {
		mockBits := []byte("hello world!")

		DBFile := &entities.File{
			ID:   primitive.NewObjectID(),
			UUID: uuid.NewString(),
			AvailableIn: []string{
				"cdn.com",
			},
			IsDeletable: false,
			Bucket:      bucket.Name,
			MimeType:    "text/plain; charset=utf-8",
			Extension:   ".txt",
		}

		url := fmt.Sprintf("https://cdn.com/%s/%s", bucket.Name, fileID /* uuid */)
		r, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		service.EXPECT().GetFileDB(gomock.Any(), bucket.Name, fileID /* uuid */).Return(DBFile, nil).Times(1)

		service.EXPECT().ReadFile(gomock.Any(), DBFile.AvailableIn).Return(mockBits, nil).Times(1)

		w := httptest.NewRecorder()

		// Will call handler.Get
		router.ServeHTTP(w, r)

		respBits := w.Body.Bytes()
		code := w.Result().StatusCode
		contentType := w.Header().Get("Content-Type")

		require.Equal(t, mockBits, respBits)
		require.Equal(t, http.StatusOK, code)
		require.Equal(t, DBFile.MimeType, contentType)

	})

	t.Run("should not get original file. File does not exist in DB. Must return 404 error", func(t *testing.T) {
		url := fmt.Sprintf("https://cdn.com/%s/%s", bucket.Name, fileID /* uuid */)
		r, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		// Returns ErrFileNotFound
		service.EXPECT().GetFileDB(gomock.Any(), bucket.Name, fileID /* uuid */).Return(nil, entities.ErrFileNotFound).Times(1)

		service.EXPECT().TryDeleteLocally(gomock.Any()).Times(1)

		w := httptest.NewRecorder()

		// Will call handler.Get
		router.ServeHTTP(w, r)

		respStr := w.Body.String()
		code := w.Result().StatusCode

		expectedResponse := fmt.Sprintf(`{"message":"%s"}`, entities.ErrFileNotFound.Error())
		require.Equal(t, expectedResponse, respStr)
		require.Equal(t, http.StatusNotFound, code)
	})

	t.Run("should get processed file that already exists", func(t *testing.T) {
		// Content-Type is text/plain; charset=utf-
		mockBits := []byte("hello world!")

		// Try get existing file bits
		service.EXPECT().ReadExisting(gomock.Any()).Return(mockBits, true /* isAvailable */, nil).Times(1)

		// Get mime type

		// Make url with query so that isOriginal inside handler is false
		url := fmt.Sprintf("https://cdn.com/%s/%s?image.resized=true", bucket.Name, fileID /* uuid */)

		r, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()

		// Will call handler.Get
		router.ServeHTTP(w, r)

		respBits := w.Body.Bytes()
		code := w.Code
		contentType := w.Header().Get("Content-Type")

		require.Equal(t, mockBits, respBits)
		require.Equal(t, http.StatusOK, code)
		require.Equal(t, "text/plain; charset=utf-8", contentType)
	})

	t.Run("should get proccessed file that does not exist", func(t *testing.T) {
		mockBits := []byte("Hello world!")
		mockResolvedBits := []byte("Hello mama!")

		DBFile := &entities.File{
			ID:   primitive.NewObjectID(),
			UUID: uuid.NewString(),
			AvailableIn: []string{
				"cdn.com",
			},
			IsDeletable: false,
			Bucket:      bucket.Name,
			MimeType:    "text/plain; charset=utf-8",
			Extension:   ".txt",
		}

		// Make it that ReadExisting returns that file is not available
		service.EXPECT().ReadExisting(gomock.Any()).Return(nil /* bits */, false /* isAvailable */, nil).Times(1)

		service.EXPECT().GetFileDB(gomock.Any(), bucket.Name, fileID /* uuid */).Return(DBFile, nil).Times(1)

		service.EXPECT().ReadFile(gomock.Any(), DBFile.AvailableIn).Return(mockBits, nil).Times(1)

		// Should be called with resolved bits
		service.EXPECT().MustSave(mockResolvedBits, gomock.Any() /* path */).Times(1)

		moduleController.EXPECT().UseResolvers(gomock.Any(), bucket.Module, gomock.Any()).DoAndReturn(
			func(buff *bytes.Buffer, module string, mm modules.ModuleMap) error {
				// Write some data to buffer. See cdn_handler.go:217
				buff.Reset()
				buff.Write(mockResolvedBits)

				return nil
			},
		).Times(1)

		// Make url with query so that isOriginal inside handler is false
		url := fmt.Sprintf("https://cdn.com/%s/%s?image.resized=true&image.webp=true", bucket.Name, fileID /* uuid */)

		r, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()

		// Will call handler.Get
		router.ServeHTTP(w, r)

		respBits := w.Body.Bytes()
		code := w.Code
		contentType := w.Header().Get("Content-Type")

		// Should be mockResolvedBits because UseResolvers has called
		require.Equal(t, mockResolvedBits, respBits)

		require.Equal(t, http.StatusOK, code)
		require.Equal(t, "text/plain; charset=utf-8", contentType)
	})

	t.Run("should return ErrNotFound because file is marked for deletion. Get original file", func(t *testing.T) {

		DBFile := &entities.File{
			ID:   primitive.NewObjectID(),
			UUID: uuid.NewString(),
			AvailableIn: []string{
				"cdn.com",
			},
			Bucket: bucket.Name,
			// Marked for deletion
			IsDeletable: true,
			MimeType:    "text/plain; charset=utf-8",
			Extension:   ".txt",
		}

		// Should return ErrFileNotFound because f.IsDeletable = true
		service.EXPECT().GetFileDB(gomock.Any(), bucket.Name, fileID /* uuid */).Return(DBFile, entities.ErrFileNotFound).Times(1)
		service.EXPECT().TryDeleteLocally(gomock.Any()).Times(1)

		// Make url with query so that isOriginal inside handler is true
		url := fmt.Sprintf("https://cdn.com/%s/%s", bucket.Name, fileID /* uuid */)

		r, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()

		// Will call handler.Get
		router.ServeHTTP(w, r)

		respStr := w.Body.String()
		code := w.Code

		expectedResponse := fmt.Sprintf(`{"message":"%s"}`, entities.ErrFileNotFound.Error())

		require.Equal(t, http.StatusNotFound, code)
		require.Equal(t, expectedResponse, respStr)
	})

}

func TestDelete(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service, moduleController := getMocks(ctrl)
	deps := setupDeps()

	handler := cdn.NewHandler(&cdn.HandlerDeps{
		Logger:           deps.Logger,
		Mux:              deps.Mux,
		BucketCache:      deps.BucketCache,
		FileCache:        deps.FileCache,
		Service:          service,
		ModuleController: moduleController,
		// Pass nil: see cdn_handler_test.go:97
		Middlewares: nil,
		MemConfig:   nil,
	})

	router := deps.Mux

	router.HandleFunc("/{bucket}/{fileUUID}", handler.Delete)

	fileID := uuid.NewString()

	t.Run("should mark file as deletable", func(t *testing.T) {
		DBFile := &entities.File{
			ID:          primitive.NewObjectID(),
			UUID:        uuid.NewString(),
			Bucket:      bucket.Name,
			AvailableIn: []string{"cdn.com"},
			MimeType:    "text/plain; charset=utf-8",
			Extension:   ".txt",
			// Not marked yet
			IsDeletable: false,
		}

		service.EXPECT().GetFileDB(gomock.Any(), bucket.Name, fileID).Return(DBFile, nil).Times(1)

		service.EXPECT().MarkAsDeletableDB(gomock.Any(), bucket.Name, DBFile.ID).Return(nil).Times(1)

		url := fmt.Sprintf("https://cdn.com/%s/%s", bucket.Name, fileID /* uuid */)

		r, err := http.NewRequest(http.MethodDelete, url, nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()

		// Will call handler.Delete
		router.ServeHTTP(w, r)

		code := w.Code

		require.Equal(t, http.StatusOK, code)
		require.Nil(t, w.Body.Bytes())
	})

	t.Run("should return error ErrFileAlreadyDeleted", func(t *testing.T) {
		DBFile := &entities.File{
			ID:          primitive.NewObjectID(),
			UUID:        uuid.NewString(),
			Bucket:      bucket.Name,
			AvailableIn: []string{"cdn.com"},
			MimeType:    "text/plain; charset=utf-8",
			Extension:   ".txt",
			// Already marked for deletion
			IsDeletable: true,
		}

		service.EXPECT().GetFileDB(gomock.Any(), bucket.Name, fileID).Return(DBFile, nil)

		url := fmt.Sprintf("https://cdn.com/%s/%s", bucket.Name, fileID /* uuid */)

		r, err := http.NewRequest(http.MethodDelete, url, nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()

		// Will call handler.Delete
		router.ServeHTTP(w, r)

		code := w.Code
		respStr := w.Body.String()

		expectedResponse := fmt.Sprintf(`{"message":"%s"}`, entities.ErrFileAlreadyDeleted.Error())

		require.Equal(t, http.StatusBadRequest, code)
		require.Equal(t, expectedResponse, respStr)
	})
}

func getMocks(ctrl *gomock.Controller) (service *mock_cdn.MockService, moduleControllerMock *mock_modules.MockController) {
	service = mock_cdn.NewMockService(ctrl)
	moduleControllerMock = mock_modules.NewMockController(ctrl)
	moduleController := modules.NewController(setupDeps().Logger)

	// Keep real implementations
	{
		moduleControllerMock.EXPECT().Raw(gomock.Any(), gomock.Any()).DoAndReturn(
			func(mm modules.ModuleMap, uuid string) string {
				return moduleController.Raw(mm, uuid)
			},
		).AnyTimes()

		moduleControllerMock.EXPECT().Parse(gomock.Any(), gomock.Any()).DoAndReturn(
			func(q url.Values, module string) (modules.ModuleMap, error) {
				return moduleController.Parse(q, module)
			},
		).AnyTimes()

		service.EXPECT().ParseMime(gomock.Any()).DoAndReturn(
			func(buff []byte) string {
				return mimetype.Detect(buff).String()
			},
		).AnyTimes()
	}

	// Naked return!
	return
}

func setupDeps() *cdn.HandlerDeps {
	logger := zap.NewNop().Sugar()
	bucketCache := bucketcache.NewBucketCache()
	fileCache := &filecache.NoOpFilecache{}
	router := mux.NewRouter()

	bucketCache.Add(bucket)

	return &cdn.HandlerDeps{
		Logger:      logger,
		BucketCache: bucketCache,
		FileCache:   fileCache,
		Mux:         router,
		Middlewares: nil,
		MemConfig:   nil,
	}

}
