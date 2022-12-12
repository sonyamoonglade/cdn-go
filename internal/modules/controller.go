package modules

import (
	module_errors "animakuro/cdn/internal/modules/errors"
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"sort"

	"go.uber.org/zap"
)

type Controller interface {
	// Applies resolvers agains buff (file processing)
	UseResolvers(buff *bytes.Buffer, module string, mm ModuleMap) error
	// Parses URL query and module into ModuleMap
	Parse(q url.Values, bucketModule string) (ModuleMap, error)
	// Sorts moduleMap and concatenates all members and uuid
	Raw(mm ModuleMap, uuid string) string
	// Checks whether module exists
	DoesModuleExist(module string) bool
}

type controller struct {
	modules map[string]*Module
	logger  *zap.SugaredLogger
}

func NewController(logger *zap.SugaredLogger) Controller {
	c := &controller{
		modules: make(map[string]*Module, 1),
		logger:  logger,
	}

	imgModule := newImageModule()
	c.registerModule(imgModule)
	return c
}

func (c *controller) Parse(q url.Values, bucketModule string) (ModuleMap, error) {
	// bucketModule represents module that bucket was created with.
	// If it doesn't exist but query has some module-related keys then return err
	if bucketModule == "" {
		return nil, module_errors.NewHttp(http.StatusBadRequest, module_errors.UnableToApplyModules)
	}

	// Get rid of auth query key
	clearQuery(&q)

	// Nothing to parse
	if len(q) == 0 {
		return nil, nil
	}

	modmap := make(ModuleMap)

	// Try to get module
	module, ok := c.modules[bucketModule]
	if !ok {
		return nil, module_errors.Wrap(ErrNotFound, http.StatusBadRequest, ErrNotFound.Error())
	}

	// Module defaults
	defaults := module.Defaults

	for defRName, defRValue := range defaults {
		modmap[defRName] = defRValue
	}

	// Number of times default value has changed to query's.
	// e.g. If client sends all default module values then
	// diffHits should be equal 0 and then nil map should be returned.
	var diffHits int

	// Allowed resolver arguments for certain module
	var allowedArguments []string
	for key, values := range q {

		resolverName, resolverArgument, module, err := valuesFromQueryPair(key, values)
		if err != nil {
			return nil, module_errors.Wrap(err, http.StatusBadRequest, err.Error())
		}

		if module != bucketModule {
			return nil, module_errors.NewHttp(http.StatusBadRequest, module_errors.ModuleNotFound, module)
		}

		allowedArguments = c.allowedArguments(module, resolverName)

		// Validate resolver value
		var ok bool
		for _, arg := range allowedArguments {
			if resolverArgument == arg {
				ok = true
				break
			}
		}

		// Invalid resolver argument is passed
		if ok == false {
			return nil, module_errors.NewHttp(http.StatusBadRequest, module_errors.UnknownResolverArgument, resolverArgument, resolverName)
		}

		// Fill map if value is not default
		if defaults[resolverName] != resolverArgument {
			modmap[resolverName] = resolverArgument
			diffHits += 1
		}

	}

	// As stated above, return nil map to indicate original file that
	// all resolver args passed are false... -> original file
	if diffHits == 0 {
		return nil, nil
	}

	return modmap, nil
}

// UseResolvers mutates initial buff according to moduleMap
func (c *controller) UseResolvers(buff *bytes.Buffer, module string, mm ModuleMap) error {
	// Prevents null check in the loop (compiler optimization)
	_ = buff
	for resolverName, resolverArg := range mm {
		r := c.resolver(module, resolverName)
		c.logger.Debugf("applying module:[%s] resolver:[%s] arg:[%s]. BuffLen: %d", module, resolverName, resolverArg, buff.Len())
		err := r(buff, resolverArg)
		if err != nil {
			return module_errors.WrapInternal(err, "controller.UseResolvers.r")
		}
	}

	return nil
}

func (c *controller) DoesModuleExist(m string) bool {
	// Empty return
	if m == "" {
		return false
	}

	_, ok := c.modules[m]

	return ok
}

func (c *controller) Raw(mm ModuleMap, uuid string) string {
	var names []string
	for k := range mm {
		names = append(names, k)
	}

	sort.Slice(names, func(i int, j int) bool {
		return names[i] < names[j]
	})

	var rawv string
	for _, resolverName := range names {
		rawv += fmt.Sprintf("%s=%s", resolverName, mm[resolverName])
	}

	return rawv + uuid
}

func (c *controller) registerModule(module *Module) {
	c.modules[module.Name] = module
}

func (c *controller) numResolvers(module string) int {
	return len(c.modules[module].Resolvers)
}

func (c *controller) resolver(module, resolverName string) ResolverFunc {
	return c.modules[module].Resolvers[resolverName]
}

func (c *controller) allowedArguments(module, resolverName string) []string {
	return c.modules[module].AllowedResolverArguments[resolverName]
}
