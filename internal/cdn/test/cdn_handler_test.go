package cdn

import (
	"testing"

	"animakuro/cdn/internal/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var bucket = &entities.Bucket{
	ID:   primitive.ObjectID{},
	Name: "site-content",
	Operations: []entities.Operation{
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
	Module: "images",
}

func TestGet(t *testing.T) {
	//
	//ctrl := gomock.NewController(t)
	//defer ctrl.Finish()
	//
	//// Deps
	//logger := zap.NewNop().Sugar()
	//bucketCache := bucketcache.NewBucketCache()
	//fileCache := &filecache.NoOpFilecache{}
	//router := mux.NewRouter()
	//service := mock_cdn.NewMockService(ctrl)
	//
	//handler := cdn.NewHandler(logger, router, service, nil /* memory config */, nil /* middlewares */, bucketCache, fileCache)
	//// Do not call handler.InitRoutes() to prevent middlewares handling the request first!
	//// That's why we pass nil
	//// handler.InitRoutes()
	//router.HandleFunc(fmt.Sprintf("/{%s}/{%s}", cdn_go.BucketKey, cdn_go.FileUUIDKey), handler.Get)
	//
	//// build path to resource
	//bucket := "site-content"
	//fileID := uuid.NewString()

	// Should serve original file that exist
	t.Run("get original file", func(t *testing.T) {
		//url := fmt.Sprintf("https://cdn.com/%s/%s", bucket, fileID)
		//
		//w := httptest.NewRecorder()
		//r, err := http.NewRequest(http.MethodGet, url, nil)
		//require.NoError(t, err)
		//
		//// Will call handler.Get
		//router.ServeHTTP(w, r)
		//
		//mockBits := []byte("hello world!")
		//
		//// Should be called one time
		//service.EXPECT().TryReadExisting("").Return(mockBits, true /* isAvailable */, nil).Times(1)
		//
		//body := w.Body.Bytes()
		//code := w.Result().StatusCode
		//
		//require.Equal(t, mockBits, body)
		//require.Equal(t, http.StatusOK, code)

	})
}
