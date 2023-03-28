package modules

import (
	module_errors "animakuro/cdn/internal/modules/errors"
	"go.uber.org/zap"

	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestModuleController(t *testing.T) {
	logger := zap.NewNop().Sugar()
	t.Run("should parse ok", func(t *testing.T) {
		t.Parallel()

		controller := NewController(logger)

		mockQuery := fmt.Sprintf("%s.webp=true&%s.resized=false", imageModuleName, imageModuleName)

		q, err := url.ParseQuery(mockQuery)
		require.NoError(t, err)

		moduleMap, err := controller.Parse(q, imageModuleName)
		require.NoError(t, err)

		// Must not be nil because query contains non false arguments
		require.NotNil(t, moduleMap)

		t.Logf("%s\n", moduleMap[imageModuleName])
	})

	t.Run("should not parse (module does not exist)", func(t *testing.T) {
		t.Parallel()

		controller := NewController(logger)

		mockQuery := "joe-biden.resolver=true"

		q, err := url.ParseQuery(mockQuery)
		require.NoError(t, err)

		moduleMap, err := controller.Parse(q, "joe-biden")
		require.Error(t, err)
		require.Nil(t, moduleMap)

		var m *module_errors.ModuleError
		var flag bool

		if errors.As(err, &m) {
			flag = true

			msg, code := m.ToHTTP()
			require.Equal(t, ErrNotFound.Error(), msg)
			require.Equal(t, http.StatusBadRequest, code)
		}

		require.True(t, flag)

	})

	t.Run("should run concurrent operations", func(t *testing.T) {
		t.Parallel()

		c := NewController(logger)

		N := 500
		wg := new(sync.WaitGroup)
		wg.Add(N)

		start := make(chan struct{})

		// Concurrent access to read-only map in modules.go
		for i := 0; i < N; i++ {
			go func() {
				<-start
				mockQuery := "image.webp=true"
				q, _ := url.ParseQuery(mockQuery)

				moduleMap, err := c.Parse(q, imageModuleName)
				require.NoError(t, err)
				require.NotNil(t, moduleMap)
				defer wg.Done()
			}()
			go func() {
				<-start
				moduleController := c.(*controller)
				moduleController.numResolvers(imageModuleName)
			}()
			go func() {
				<-start
				moduleController := c.(*controller)
				moduleController.resolver(imageModuleName, webp)
			}()
		}
		time.Sleep(time.Millisecond * 50)

		// Send signal to all goroutines
		close(start)

		wg.Wait()

	})
}
