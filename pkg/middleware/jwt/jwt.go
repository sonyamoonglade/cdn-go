package jwt

import (
	"net/http"
	"strings"

	cdn_go "animakuro/cdn"
	cache "animakuro/cdn/pkg/cache/bucket"
	"animakuro/cdn/pkg/cdn_errors"
	"github.com/cristalhq/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var ErrMissingAuthHeader = errors.New("missing authorization header")
var ErrAccessDenied = errors.New("access denied")
var ErrMissingAuthKey = errors.New("missing auth key in url")

type Middleware struct {
	logger *zap.SugaredLogger
	bc     *cache.BucketCache
}

type Claims struct {
	Bucket string `json:"bucket"`
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

		m.logger.Debugf("operation: %s bucket: %s", operation, bucketName)

		b, err := m.bc.Get(bucketName)
		if err != nil {
			cdn_errors.ToHttp(m.logger, w, err)
			return
		}

		var keys []string
		for _, op := range b.Operations {
			if op.Name == operation {
				//No jwt verification if operation is public
				if op.Type == cdn_go.OperationTypePublic {
					h.ServeHTTP(w, r)
					return
				}
				keys = op.Keys
			}
		}

		var token []byte

		if operation == cdn_go.OperationGet {
			key := r.URL.Query().Get(cdn_go.URLAuthKey)
			if key == "" {
				cdn_errors.ToHttp(m.logger, w, ErrMissingAuthKey)
				return
			}

			token = []byte(key)

		} else {
			bearer := r.Header.Get("Authorization")
			if bearer == "" {
				cdn_errors.ToHttp(m.logger, w, ErrMissingAuthHeader)
				return
			}

			token = []byte(strings.Split(bearer, " ")[1])
		}

		var vrf *jwt.HSAlg
		var claims Claims
		var ok bool

		for _, key := range keys {
			//Ignore error (alg is always supported, key is neven empty or nil). See impl. of jwt.NewVerifierHS
			vrf, _ = jwt.NewVerifierHS(jwt.HS256, []byte(key))
			err := jwt.ParseClaims(token, vrf, &claims)
			if err != nil {
				//Skip invalid signature to check all keys
				if errors.Is(err, jwt.ErrInvalidSignature) {
					m.logger.Debugf("invalid token signature. Retrying")
					continue
				}
				if errors.Is(err, jwt.ErrInvalidFormat) {
					m.logger.Debugf("invalid token format. Retrying")
					continue
				}

				if errors.Is(err, jwt.ErrInvalidKey) {
					m.logger.Debugf("invalid token key. Retrying")
					continue
				}

				err = cdn_errors.WrapInternal(err, "JwtMiddleware.Auth.jwt.ParseClaims")
				cdn_errors.ToHttp(m.logger, w, err)
				return
			}
			//Verification successful
			ok = true
			m.logger.Debugf("token is valid")
		}

		//Handle invalid jwt
		if ok == false {
			cdn_errors.ToHttp(m.logger, w, ErrAccessDenied)
			return
		}

		//Jwt is valid
		h.ServeHTTP(w, r)
	}
}
