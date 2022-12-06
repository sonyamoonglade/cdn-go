package auth

import (
	"net/url"
	"strings"

	cdn_go "animakuro/cdn"
	"animakuro/cdn/internal/cdn/cdnutil"

	"github.com/cristalhq/jwt/v4"
	"github.com/pkg/errors"
)

type Claims struct {
	// Requested bucket
	Bucket string `json:"bucket"`

	// Requested file uuid
	FileID string `json:"file_id"`
}

var (
	ErrMissingAuthHeader = errors.New("missing authorization header")
	ErrInvalidAuthHeader = errors.New("invalid authorization header")
	ErrAccessDenied      = errors.New("access denied")
	ErrMissingAuthKey    = errors.New("missing auth key in url")
)

// ParseToken tries to parse a []byte token from url or tokenSource according to operation
func ParseToken(operation string, url *url.URL, tokenSource string) ([]byte, error) {
	var token []byte

	if operation == cdn_go.OperationGet {
		key := url.Query().Get(cdn_go.URLAuthKey)
		if key == "" {
			return nil, ErrMissingAuthKey
		}

		token = []byte(key)

	} else {
		// tokenSource could be an authorization header
		if tokenSource == "" {
			return nil, ErrMissingAuthHeader

		}

		splitBySpace := strings.Split(tokenSource, " ")
		if len(splitBySpace) == 1 {
			return nil, ErrInvalidAuthHeader
		}

		token = []byte(splitBySpace[1])
	}

	return token, nil
}

func ValidateToken(token []byte, keys []string, wanted *Claims) (bool, error) {
	var vrf *jwt.HSAlg
	var claims Claims
	var ok bool
	for _, key := range keys {
		// Ignore error (alg is always supported, key is neven empty or nil). See impl. of jwt.NewVerifierHS
		vrf, _ = jwt.NewVerifierHS(jwt.HS256, []byte(key))
		err := jwt.ParseClaims(token, vrf, &claims)
		if err != nil {
			// Skip in order to check all keys
			if errors.Is(err, jwt.ErrInvalidSignature) {
				continue
			}

			if errors.Is(err, jwt.ErrInvalidFormat) {
				continue
			}

			if errors.Is(err, jwt.ErrInvalidKey) {
				continue
			}

			return false, cdnutil.WrapInternal(err, "auth.ValidateToken.jwt.ParseClaims")
		}

		// Verification successful
		// Token and payload is correct
		if wanted.FileID == claims.FileID && wanted.Bucket == claims.Bucket {
			ok = true
		}

	}

	return ok, nil
}
