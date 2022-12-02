package modules

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	cdn_go "animakuro/cdn"
	module_errors "animakuro/cdn/internal/modules/errors"
	"animakuro/cdn/internal/modules/image"
	"animakuro/cdn/internal/modules/types"
)

var (
	modules = make(map[string]*types.Module)
)

var (
	ErrInvalidURL = errors.New("invalid url")
)

// Init calls .Init function of all cdn modules
func Init() {
	img := image.Init()
	modules[image.ModuleName] = img
}

func NumResolvers(module string) int {
	return len(modules[module].Resolvers)
}

func Resolver(module, resolverName string) types.ResolverFunc {
	return modules[module].Resolvers[resolverName]
}

func Parse(q url.Values, bucketModule string) (types.ModuleMap, error) {

	//bucketModule represents module that bucket was created with
	//if it doesn't present but query has some module-related keys -> err
	if bucketModule == "" {
		return nil, module_errors.NewHttp(http.StatusBadRequest, module_errors.UnableToApplyModules)
	}

	clearQuery(&q)

	if len(q) == 0 {
		return nil, nil
	}

	modmap := make(types.ModuleMap)

	//module defaults
	defaults := modules[bucketModule].Defaults

	for defRName, defRValue := range defaults {
		modmap[defRName] = defRValue
	}

	//number of times default value has changed to query's
	//e.g. if client sends all default module values
	//diffHits should be 0 and then nil map should be returned.
	var diffHits int

	//module allowed resolver values
	var allowedModuleRV []string
	for k, vv := range q {

		resolverName, resolverv, queryModule, err := extractUrlValues(k, vv)

		if err != nil {
			return nil, module_errors.Wrap(err, http.StatusBadRequest, err.Error())
		}

		if queryModule != bucketModule {
			return nil, module_errors.NewHttp(http.StatusBadRequest, module_errors.ModuleNotFound, queryModule)
		}

		//Validate resolver value
		allowedModuleRV = modules[bucketModule].AllowedResolverValues[resolverName]

		var ok bool
		for _, allowedRV := range allowedModuleRV {
			if resolverv == allowedRV {
				ok = true
				break
			}
		}

		//Invalid resolver value passed
		if ok == false {
			return nil, module_errors.NewHttp(http.StatusBadRequest, module_errors.UnknownResolverArgument, resolverv, resolverName)
		}

		//Fill map if value is not default
		if defaults[resolverName] != resolverv {
			modmap[resolverName] = resolverv
			diffHits += 1
		}

	}

	//As stated above, return nil map to indicate original file that
	//all resolver args passed are false... -> original file
	if diffHits == 0 {
		return nil, nil
	}

	return modmap, nil
}

// UseResolvers mutates initial buff according to moduleMap
func UseResolvers(buff *bytes.Buffer, module string, mm types.ModuleMap) error {
	// Prevents null check in the loop (compiler optimization)
	_ = buff
	for resolverName, resolverArg := range mm {
		r := Resolver(module, resolverName)
		err := r(buff, resolverArg)
		if err != nil {
			return err
		}
	}

	return nil
}

func DoesModuleExist(m string) error {
	//no modules are being applied
	if m == "" {
		return nil
	}
	_, ok := modules[m]
	if ok == false {
		return module_errors.NewHttp(http.StatusBadRequest, module_errors.ModuleNotFound, m)
	}

	return nil
}

// Raw takes moduleMap and uuid of file and concatenates it to one string
// Internally does sort the moduleMap
func Raw(modmap types.ModuleMap, uuid string) string {
	var names []string
	for k := range modmap {
		names = append(names, k)
	}

	sort.Slice(names, func(i int, j int) bool {
		return names[i] < names[j]
	})

	var rawv string
	for _, resolverName := range names {
		rawv += fmt.Sprintf("%s=%s", resolverName, modmap[resolverName])
	}

	return rawv + uuid
}

func extractUrlValues(qkey string, qval []string) (string, string, string, error) {

	ksplit := strings.Split(qkey, ".")

	if ksplit[0] == cdn_go.URLAuthKey {
		return "", "", "", nil
	}

	if len(ksplit) == 1 {
		return "", "", "", ErrInvalidURL
	}

	resolverName := ksplit[1]
	resolverValue := qval[0]
	queryModule := ksplit[0]

	return resolverName, resolverValue, queryModule, nil
}

//clearQuery removes all unnecessary query keys for module parsing
func clearQuery(q *url.Values) {
	q.Del(cdn_go.URLAuthKey)
}
