package jwt

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"animakuro/cdn/internal/auth"
	"animakuro/cdn/internal/entities"
	cache "animakuro/cdn/pkg/cache/bucket"

	"github.com/cristalhq/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

var bucket = &entities.Bucket{
	ID:   primitive.ObjectID{},
	Name: "images",
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
		// Keeping keys as nil (empty slice) and operation to private.
		// This will make delete operation in this bucket completely unreachable.
		{
			Name: "delete",
			Type: "private",
			Keys: nil,
		},
	},
	Module: "images",
}

func TestAuth(t *testing.T) {
	bc := cache.NewBucketCache()
	bc.Add(bucket)

	m := NewMiddleware(zap.NewNop().Sugar(), bc)

	authfn := m.Auth(func(w http.ResponseWriter, r *http.Request) {
		// mock
		w.WriteHeader(http.StatusOK)
	})

	// Setup router to use mux.Vars
	router := mux.NewRouter()
	router.Handle("/{bucket}/{fileUUID}", authfn)
	router.Handle("/{bucket}", authfn)

	t.Run("should allow private get", func(t *testing.T) {
		t.Parallel()

		// Requested file
		fileID := "abcd-efgh"

		// Setup token
		// Key is same as in bucket defined in bucketCache
		key := "abcd"
		signer, _ := jwt.NewSignerHS(jwt.HS256, []byte(key))
		builder := jwt.NewBuilder(signer)

		payload := auth.Claims{
			Bucket: bucket.Name,
			FileID: fileID,
		}

		token, err := builder.Build(payload)
		require.NoError(t, err)

		parsedUrl, err := url.Parse(fmt.Sprintf("https://cdn.com/%s/%s?auth=%s", bucket.Name, fileID, token))
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, parsedUrl.String(), nil)
		require.NoError(t, err)

		router.ServeHTTP(w, req)

		respBody := w.Body.String()
		status := w.Code

		// Means auth handler authed the request and let handler defined above return http.StatusOK
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, "", respBody)
	})

	t.Run("should deny private get. Invalid fileID in jwt payload", func(t *testing.T) {
		t.Parallel()
		// Requested file (NOT SAME AS IN PAYLOAD)
		fileID := "random-bullshit"

		// Setup token
		// Key is same as in bucket defined in bucketCache
		key := "abcd"
		signer, _ := jwt.NewSignerHS(jwt.HS256, []byte(key))
		builder := jwt.NewBuilder(signer)

		payload := auth.Claims{
			Bucket: bucket.Name,
			FileID: "some different file id from requested",
		}

		token, err := builder.Build(payload)
		require.NoError(t, err)

		parsedUrl, err := url.Parse(fmt.Sprintf("https://cdn.com/%s/%s?auth=%s", bucket.Name, fileID, token))
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, parsedUrl.String(), nil)
		require.NoError(t, err)

		router.ServeHTTP(w, req)

		respBody := w.Body.String()
		status := w.Code

		// In case requested fileID differs from payload it should deny acecss
		require.Equal(t, http.StatusForbidden, status)

		expectedResponse := `{"message":"access denied"}`
		require.Equal(t, expectedResponse, respBody)
	})

	t.Run("should allow private post", func(t *testing.T) {
		t.Parallel()

		// Setup token
		// Key is same as in bucket defined in bucketCache
		key := "abcd"
		signer, _ := jwt.NewSignerHS(jwt.HS256, []byte(key))
		builder := jwt.NewBuilder(signer)

		payload := auth.Claims{
			Bucket: bucket.Name,
			// Not providing FileID for upload operation
			FileID: "",
		}

		token, err := builder.Build(payload)
		require.NoError(t, err)

		uploadURL := fmt.Sprintf("https://cdn.com/%s", bucket.Name)

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, uploadURL, nil)
		require.NoError(t, err)

		// Set auth header
		req.Header.Set("Authorization", "Bearer "+token.String())

		router.ServeHTTP(w, req)

		status := w.Code

		require.Equal(t, http.StatusOK, status)
		require.Zero(t, w.Body.Len())
	})

	t.Run("should deny private upload. Invalid bucket in jwt payload", func(t *testing.T) {
		t.Parallel()

		// Setup token
		// Key is same as in bucket defined in bucketCache
		key := "abcd"
		signer, _ := jwt.NewSignerHS(jwt.HS256, []byte(key))
		builder := jwt.NewBuilder(signer)

		payload := auth.Claims{
			// Not equal to bucket.Name !!
			Bucket: "random-bullshit",
			// Not providing FileID for upload operation
			FileID: "",
		}

		token, err := builder.Build(payload)
		require.NoError(t, err)

		uploadURL := fmt.Sprintf("https://cdn.com/%s", bucket.Name)

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, uploadURL, nil)
		require.NoError(t, err)

		// Set auth header
		req.Header.Set("Authorization", "Bearer "+token.String())

		router.ServeHTTP(w, req)

		status := w.Code
		respBody := w.Body.String()

		expectedResponse := `{"message":"access denied"}`
		require.Equal(t, expectedResponse, respBody)
		require.Equal(t, http.StatusForbidden, status)
	})

	t.Run("should deny access to private delete. Empty keys and private operation", func(t *testing.T) {
		t.Parallel()

		// FileID to delete
		fileID := "abcd-efgh-1234"

		// Setup token
		// Key is same as in bucket defined in bucketCache
		key := "abcd"
		signer, _ := jwt.NewSignerHS(jwt.HS256, []byte(key))
		builder := jwt.NewBuilder(signer)

		payload := auth.Claims{
			Bucket: bucket.Name,
			// Not providing FileID for upload operation
			FileID: fileID,
		}

		token, err := builder.Build(payload)
		require.NoError(t, err)

		deleteURL := fmt.Sprintf("https://cdn.com/%s/%s", bucket.Name, fileID)

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
		require.NoError(t, err)

		// Set auth header
		req.Header.Set("Authorization", "Bearer "+token.String())

		router.ServeHTTP(w, req)

		status := w.Code
		respBody := w.Body.String()

		// Denies access. See jwt.go:63
		expectedResponse := `{"message":"access denied"}`
		require.Equal(t, expectedResponse, respBody)
		require.Equal(t, http.StatusForbidden, status)
	})

	// todo: missing tokens etc..
}
