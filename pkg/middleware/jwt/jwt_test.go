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

	t.Run("private get OK", func(t *testing.T) {
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

		parsedUrl, err := url.Parse(fmt.Sprintf("https://cdn.com/images/%s?auth=%s", fileID, token))
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, parsedUrl.String(), nil)
		require.NoError(t, err)

		router.ServeHTTP(w, req)

		respBody := w.Body.String()
		status := w.Code
		fmt.Println(respBody, status)
		// Means auth handler authed the request and let handler defined above return http.StatusOK
		require.Equal(t, http.StatusOK, status)
		require.Equal(t, "", respBody)
	})

	t.Run("private get. Invalid fileID in payload", func(t *testing.T) {
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

		parsedUrl, err := url.Parse(fmt.Sprintf("https://cdn.com/images/%s?auth=%s", fileID, token))
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

	// todo: missing tokens etc..
}
