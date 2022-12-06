package response_test

import (
	"animakuro/cdn/pkg/http/response"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInternal(t *testing.T) {

	w := httptest.NewRecorder()

	response.Internal(w)

	expectedBody := "Internal error"
	expectedCode := 500

	actualBody, err := io.ReadAll(w.Body)

	require.NoError(t, err)
	require.Equal(t, expectedBody, string(actualBody))
	require.Equal(t, expectedCode, w.Code)

}

func TestJson(t *testing.T) {

	prodLogger, err := zap.NewProduction()
	require.NoError(t, err)
	logger := prodLogger.Sugar()

	w := httptest.NewRecorder()

	content := map[string]interface{}{
		"message": "some testing message",
	}
	expectedContent, _ := json.Marshal(content)

	response.Json(logger, w, http.StatusOK, content)

	respBytes, err := io.ReadAll(w.Body)
	require.NoError(t, err)

	require.Equal(t, expectedContent, respBytes)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, w.Header().Get("Content-Type"), "application/json")
}
