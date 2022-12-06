package modules

import (
	"bytes"
	"errors"
	"net/url"
	"strings"

	cdn_go "animakuro/cdn"
)

var (
	ErrInvalidURL  = errors.New("invalid url")
	ErrInvalidArgs = errors.New("invalid arguments")
	ErrNotFound    = errors.New("module not found")
)

const (
	TrueStr  = "true"
	FalseStr = "false"
)

type Module struct {
	Name                     string
	Resolvers                map[string]ResolverFunc
	Defaults                 Defaults
	AllowedResolverArguments map[string][]string
}

type (
	Defaults map[string]string

	ModuleMap map[string]string

	ResolverFunc func(buff *bytes.Buffer, arg interface{}) error

	RegisterFunc func() *Module
)

// valuesFromQueryPair uses naked return in order to clarify returning params
func valuesFromQueryPair(key string, values []string) (resolverName string, resolverArgument string, module string, err error) {

	ksplit := strings.Split(key, ".")

	// Skip this key
	if ksplit[0] == cdn_go.URLAuthKey {
		return
	}

	if len(ksplit) == 1 {
		err = ErrInvalidURL
		return
	}

	resolverName = ksplit[1]
	resolverArgument = values[0]
	module = ksplit[0]

	err = nil
	return
}

//clearQuery removes all unnecessary query keys for module parsing
func clearQuery(q *url.Values) {
	q.Del(cdn_go.URLAuthKey)
}
