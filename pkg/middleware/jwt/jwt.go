package jwt

import (
	"net/http"
	"strings"

	cdn_go "animakuro/cdn"
	"animakuro/cdn/internal/auth"
	cdn_errors "animakuro/cdn/internal/cdn/errors"
	cache "animakuro/cdn/pkg/cache/bucket"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Middleware struct {
	logger *zap.SugaredLogger
	bc     *cache.BucketCache
}

func NewMiddleware(logger *zap.SugaredLogger, bucketCache *cache.BucketCache) *Middleware {
	return &Middleware{
		logger: logger,
		bc:     bucketCache,
	}
}

func (m *Middleware) Auth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)

		operation := strings.ToLower(r.Method)
		bucketName := vars[cdn_go.BucketKey]

		// TODO: rename
		var fileUUID string

		// For get, post operations FileUUIDKey is provided.
		if operation != cdn_go.OperationPost {
			fileUUID = vars[cdn_go.FileUUIDKey]
		}

		m.logger.Debugf("operation: %s bucket: %s", operation, bucketName)

		b, err := m.bc.Get(bucketName)
		if err != nil {
			cdn_errors.ToHttp(m.logger, w, err)
			return
		}

		var keys []string
		for _, op := range b.Operations {
			if op.Name == operation {
				// No jwt verification if operation is public
				if op.Type == cdn_go.OperationTypePublic {
					h.ServeHTTP(w, r)
					return
				}

				// If op.Keys is empty and operation type is private then access must be denied.
				// Omit the check for private operation. (See jwt.go:56)
				if op.Keys == nil {
					cdn_errors.ToHttp(m.logger, w, auth.ErrAccessDenied)
					return
				}

				keys = op.Keys
			}
		}

		// Get token according to operation
		token, err := auth.ParseToken(operation, r.URL, r.Header.Get("Authorization"))
		if err != nil {
			cdn_errors.ToHttp(m.logger, w, err)
			return
		}

		wantedClaims := auth.Claims{
			Bucket: bucketName,
			FileID: fileUUID,
		}

		// Validate token based on internals an wantedClaims
		ok, err := auth.ValidateToken(token, keys, &wantedClaims)
		if err != nil {
			cdn_errors.ToHttp(m.logger, w, err)
			return
		}

		//Handle invalid jwt
		if ok == false {
			cdn_errors.ToHttp(m.logger, w, auth.ErrAccessDenied)
			return
		}

		//Jwt is valid
		h.ServeHTTP(w, r)
	}
}
