package cdn_errors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	errmod "animakuro/cdn/internal/modules/errors"
	"github.com/sonyamoonglade/notification-service/pkg/response"
	"go.uber.org/zap"
)

var ErrFileAlreadyExists = errors.New("file(s) already exist")
var ErrFileNotFound = errors.New("file not found")

var ErrBucketNotFound = errors.New("bucket not found")
var ErrBucketAlreadyExists = errors.New("bucket already exists")
var ErrBucketsAreNotDefined = errors.New("no buckets are defined in database")
var ErrCouldNotRemoveFile = errors.New("could not remove file")

var ErrInvalidUrl = errors.New("invalid url")

const (
	InternalError = "internal error"
)

func WrapInternal(err error, context string) error {
	return fmt.Errorf("internal error %s: %v", context, err)
}

func ToHttp(logger *zap.SugaredLogger, w http.ResponseWriter, err error) {

	var m *errmod.ModuleError
	skiplog := false

	var resp string
	var code int

	if errors.As(err, &m) {
		//Unwrapping original wrapped error
		err = m.Unwrap()
		if err != nil {
			logger.Error(err.Error())
		}

		resp, code = m.ToHTTP()
		skiplog = true
	} else {
		//Not module error
		resp, code = parse(err)
	}

	response.Json(logger, w, code, response.JSON{
		"message": resp,
	})

	if code > http.StatusMethodNotAllowed && skiplog == false {
		logger.Error(err.Error())
	}

	return
}

func parse(err error) (string, int) {

	text := err.Error()

	//todo: introduce custom error already with client response and status code...
	switch true {
	case strings.Contains(text, "binding error"):
		return err.Error(), http.StatusBadRequest
	case strings.Contains(text, "missing auth key"):
		return err.Error(), http.StatusUnauthorized
	case strings.Contains(text, "could not remove file"):
		return err.Error(), http.StatusServiceUnavailable
	case strings.Contains(text, "missing authorization header"):
		return err.Error(), http.StatusUnauthorized
	case strings.Contains(text, "access denied"):
		return err.Error(), http.StatusForbidden
	case strings.Contains(text, "file(s) already exist"):
		return err.Error(), http.StatusBadRequest
	case strings.Contains(text, "already exists"):
		return err.Error(), http.StatusConflict
	case strings.Contains(text, "validation error"):
		return err.Error(), http.StatusBadRequest
	case strings.Contains(text, "internal error"):
		return InternalError, http.StatusInternalServerError
	case strings.Contains(text, "not found"):
		return err.Error(), http.StatusNotFound
	case strings.Contains(text, "API key is not present in the request"):
		return err.Error(), http.StatusForbidden
	case strings.Contains(text, "invalid API key"):
		return err.Error(), http.StatusForbidden
	case strings.Contains(text, "file has invalid extension"):
		return err.Error(), http.StatusBadRequest

	default:
		//TODO: add default err
		return InternalError, http.StatusInternalServerError
	}

}
