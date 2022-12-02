package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAppConfig(t *testing.T) {

	//Imitate .env file's vars
	os.Setenv("MONGO_URI", "mock_mongo_uri")
	os.Setenv("APP_PORT", "9000")
	os.Setenv("APP_HOST", "localhost")
	os.Setenv("MONGO_DB_NAME", "dbname")
	os.Setenv("DOMAIN", "mock_domain")

	cfg, err := GetAppConfig("./testdata/config.yaml", true)
	require.NoError(t, err)

	//envs..
	require.Equal(t, "mock_mongo_uri", cfg.MongoURI)
	require.Equal(t, "9000", cfg.AppPort)
	require.Equal(t, "localhost", cfg.AppHost)
	require.Equal(t, "dbname", cfg.DBName)
	require.Equal(t, true, cfg.Debug)
	require.Equal(t, "mock_domain", cfg.Domain)

	//cfg file
	require.Equal(t, int64(128), cfg.MemoryConfig.MaxUploadSize)
	require.Equal(t, 100, cfg.MaxWorkers)
	require.Equal(t, 5, cfg.FileCacheConfig.CacheTTL)
	require.Equal(t, 5, cfg.FileCacheConfig.CacheThreshold)
	require.Equal(t, 60, cfg.FileCacheConfig.CheckoutEvery)
	require.Equal(t, int64(512), cfg.FileCacheConfig.MaxCacheSize)
	require.Equal(t, 128, cfg.FileCacheConfig.MaxCacheItems)
	require.Equal(t, 120, cfg.FileCacheConfig.FlushEvery)

}
