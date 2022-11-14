package types

import "bytes"

type Module struct {
	Name                  string
	Resolvers             map[string]ResolverFunc
	Defaults              Defaults
	AllowedResolverValues map[string][]string
}

type Defaults map[string]string

type ModuleMap map[string]string

type ResolverFunc func(buff *bytes.Buffer, arg interface{}) error
