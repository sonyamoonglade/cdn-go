package cdn_errors

import (
	"errors"
	"net/http"
	"strings"

	"animakuro/cdn/internal/auth"
	"animakuro/cdn/internal/cdn/cdnutil"
	"animakuro/cdn/internal/entities"
	"animakuro/cdn/internal/formdata"
	"animakuro/cdn/internal/modules"
	module_errors "animakuro/cdn/internal/modules/errors"
	"animakuro/cdn/pkg/http/response"

	"go.uber.org/zap"
)

func ToHttp(logger *zap.SugaredLogger, w http.ResponseWriter, err error) {

	var (
		m       *module_errors.ModuleError
		skiplog bool
		resp    string
		code    int
	)

	if errors.As(err, &m) {
		// Unwrap original wrapped error
		err = m.Unwrap()
		if err != nil {
			// Log full error if it's moduleError
			logger.Error(err.Error())
		}
		resp, code = m.ToHTTP()
		logger.Debug(resp, code)
		skiplog = true
	} else {
		// Not module error
		resp, code = parse(err)
	}

	response.Json(logger, w, code, response.JSON{
		"message": resp,
	})

	if code > http.StatusMethodNotAllowed && !skiplog {
		logger.Error(err.Error())
	}

	return
}

func parse(err error) (string, int) {

	is := cdnutil.IsErrorOf(err)

	switch true {

	// TODO: remove strings.Contains
	case strings.Contains(err.Error(), "binding error"):
		return err.Error(), http.StatusBadRequest

	case strings.Contains(err.Error(), "validation"):
		return err.Error(), http.StatusBadRequest

	// Auth
	case is(auth.ErrMissingAuthHeader):
		return err.Error(), http.StatusUnauthorized

	case is(auth.ErrMissingAuthKey):
		return err.Error(), http.StatusUnauthorized

	case is(auth.ErrInvalidAuthHeader):
		return err.Error(), http.StatusUnauthorized

	case is(auth.ErrAccessDenied):
		return err.Error(), http.StatusForbidden
	// --- Auth END

	// File entity
	case is(entities.ErrFileCantRemove):
		return err.Error(), http.StatusServiceUnavailable

	case is(entities.ErrFileNotFound):
		return err.Error(), http.StatusNotFound

	case is(entities.ErrFileAlreadyExists):
		return err.Error(), http.StatusConflict
	// --- File entity END

	// Bucket entity
	case is(entities.ErrBucketNotFound):
		return err.Error(), http.StatusNotFound

	case is(entities.ErrBucketAlreadyExists):
		return err.Error(), http.StatusConflict
	// --- Bucket entity END

	// Formdata
	case is(formdata.ErrInvalidExtension):
		return err.Error(), http.StatusBadRequest

	case is(entities.ErrNoFiles):
		return err.Error(), http.StatusBadRequest
	// --- Formdata END

	// Module errors
	case is(modules.ErrNotFound):
		return err.Error(), http.StatusBadRequest
	// --- Module END

	default:
		return "Internal error", http.StatusInternalServerError
	}

}
