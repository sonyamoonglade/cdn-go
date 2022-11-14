package module_errors

import (
	"fmt"
	"net/http"
)

//Templates
var (
	ModuleNotFound          = "module %s not found"
	UnknownResolverArgument = "unknown resolver argument %s on resolver %s"
	UnableToApplyModules    = "unable to apply modules for this bucket"
)

const (
	InternalError = "internal error"
)

type ModuleError struct {
	clientMsg string
	wrapped   error
	code      int
}

func (m *ModuleError) Error() string {
	return fmt.Sprintf("module error: wraps %v", m.wrapped)
}

func (m *ModuleError) Unwrap() error {
	return m.wrapped
}

func (m *ModuleError) ToHTTP() (string, int) {
	return m.clientMsg, m.code
}

func Wrap(err error, code int, msg string, args ...interface{}) error {
	return &ModuleError{
		clientMsg: fmt.Sprintf(msg, args...),
		wrapped:   err,
		code:      code,
	}
}

func NewHttp(code int, msg string, args ...interface{}) error {
	return &ModuleError{
		clientMsg: fmt.Sprintf(msg, args...),
		wrapped:   nil,
		code:      code,
	}
}

func WrapInternal(err error, context string) error {
	return &ModuleError{
		clientMsg: InternalError,
		wrapped:   fmt.Errorf("module error: %s:%v", context, err),
		code:      http.StatusInternalServerError,
	}
}
