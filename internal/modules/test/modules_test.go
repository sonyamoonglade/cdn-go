package modules

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"

	"animakuro/cdn/internal/modules"
	module_errors "animakuro/cdn/internal/modules/errors"
	"animakuro/cdn/internal/modules/image"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

var once = new(sync.Once)

func init() {
	once.Do(modules.Init)
}

func TestParseOk(t *testing.T) {

	mockQuery := fmt.Sprintf("%s.%s=%s", image.ModuleName, image.Webp, image.TrueStr)
	q, err := url.ParseQuery(mockQuery)
	require.NoError(t, err)

	moduleMap, err := modules.Parse(q, image.ModuleName)
	require.NoError(t, err)
	require.NotNil(t, moduleMap)

	//modules.Parse should assign default values to resolvers that didn't come in URL (see modules.Parse impl.)
	//"image" module has registered n resolvers.
	n := modules.NumResolvers(image.ModuleName)

	require.Equal(t, n, len(moduleMap))

	webpv, ok := moduleMap[image.Webp]
	require.True(t, ok)
	require.Equal(t, image.TrueStr, webpv)

	resizedv, ok := moduleMap[image.Resized]
	require.True(t, ok)
	require.Equal(t, image.FalseStr, resizedv)

}

func TestParseModuleDoesNotExist(t *testing.T) {

	mockQuery := fmt.Sprintf("%s.%s=%s", "random-module", image.Webp, image.TrueStr)
	q, err := url.ParseQuery(mockQuery)
	require.NoError(t, err)

	moduleMap, err := modules.Parse(q, image.ModuleName)
	require.Error(t, err)
	require.Nil(t, moduleMap)

	var m *module_errors.ModuleError
	var flag bool

	if errors.As(err, &m) {
		flag = true
		require.True(t, flag)

		msg, code := m.ToHTTP()
		require.True(t, strings.Contains(msg, "random-module not found"))
		require.Equal(t, http.StatusBadRequest, code)
	}

}

func BenchmarkParse(b *testing.B) {

	mockQuery := fmt.Sprintf("%s.%s=%s", image.ModuleName, image.Webp, image.TrueStr)
	q, err := url.ParseQuery(mockQuery)
	if err != nil {
		b.Fatal(err.Error())

	}

	for i := 0; i < b.N; i++ {
		_, err := modules.Parse(q, image.ModuleName)
		if err != nil {
			b.Fatal(err.Error())
		}

	}
	b.ReportAllocs()

}
