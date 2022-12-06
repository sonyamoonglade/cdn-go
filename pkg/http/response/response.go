package response

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type JSON map[string]interface{}

func Ok(w http.ResponseWriter) {
	w.WriteHeader(200)
	return
}

func Created(w http.ResponseWriter) {
	w.WriteHeader(201)
	return
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(204)
	return
}

func Internal(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal error"))
	return
}

func Json(logger *zap.SugaredLogger, w http.ResponseWriter, code int, content JSON) {
	bytes, err := json.Marshal(content)
	if err != nil {
		logger.Error(err.Error())
		Internal(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	w.Write(bytes)
	return
}

func Binary(w http.ResponseWriter, buff []byte, mime string) {
	ct := "application/octet-stream"
	if mime != "" {
		ct = mime
	}

	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(buff)))
	w.WriteHeader(http.StatusOK)
	w.Write(buff)
}
