package config

import (
	"fmt"
	"os"

	filecache "animakuro/cdn/pkg/cache/file"
	"github.com/spf13/viper"
)

type MemoryConfig struct {
	MaxUploadSize int64 // Represents megabytes 10^6 byte
}

type AppConfig struct {
	MongoURI        string
	DBName          string
	AppPort         string
	AppHost         string
	Debug           bool
	Domain          string
	MaxWorkers      int
	MemoryConfig    *MemoryConfig
	FileCacheConfig *filecache.Config
}

const BasePath = "config.yaml"

func GetAppConfig(v *viper.Viper, debug bool) (*AppConfig, error) {

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		return nil, fmt.Errorf("missing MONGO_URI env")
	}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		return nil, fmt.Errorf("missing APP_PORT env")
	}

	appHost := os.Getenv("APP_HOST")
	if appHost == "" {
		return nil, fmt.Errorf("missing APP_HOST env")
	}

	dbname := os.Getenv("MONGO_DB_NAME")
	if dbname == "" {
		return nil, fmt.Errorf("missing MONGO_DB_NAME env")
	}

	env := os.Getenv("ENV")
	if env == "" {
		return nil, fmt.Errorf("missing ENV env")
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		return nil, fmt.Errorf("missing DOMAIN env")
	}

	cacheMaxMem := v.GetInt64("cache.max_memory")
	if cacheMaxMem == 0 {
		return nil, fmt.Errorf("missing cache.max_memory in config")
	}

	cacheMaxItems := v.GetInt("cache.max_items")
	if cacheMaxItems == 0 {
		return nil, fmt.Errorf("missing cache.max_items in config")
	}

	cacheTtl := v.GetInt("cache.ttl")
	if cacheTtl == 0 {
		return nil, fmt.Errorf("missing cache.ttl in config")
	}

	cacheCheckoutEvery := v.GetInt("cache.checkout_every")
	if cacheCheckoutEvery == 0 {
		return nil, fmt.Errorf("missing cache.checkout_every in config")
	}

	cacheThreshold := v.GetInt("cache.threshold")
	if cacheThreshold == 0 {
		return nil, fmt.Errorf("missing cache.threshold in config")
	}

	cacheFlushEvery := v.GetInt("cache.flush_every")
	if cacheFlushEvery == 0 {
		return nil, fmt.Errorf("missing cache.flush_every in config")
	}

	uploadMaxMem := v.GetInt64("cdn.upload.max_memory")
	if uploadMaxMem == 0 {
		return nil, fmt.Errorf("missing cdn.upload.max_memory in config")
	}

	maxWorkers := v.GetInt("cdn.io_workers")
	if maxWorkers == 0 {
		return nil, fmt.Errorf("missing cdn.io_workers in config")
	}

	return &AppConfig{
		MongoURI:   mongoURI,
		AppPort:    appPort,
		AppHost:    appHost,
		Debug:      debug,
		DBName:     dbname,
		MaxWorkers: maxWorkers,
		Domain:     domain,
		MemoryConfig: &MemoryConfig{
			MaxUploadSize: uploadMaxMem,
		},
		FileCacheConfig: &filecache.Config{
			MaxCacheSize:   cacheMaxMem,
			MaxCacheItems:  cacheMaxItems,
			CacheTTL:       cacheTtl,
			CacheThreshold: cacheThreshold,
			FlushEvery:     cacheFlushEvery,
			CheckoutEvery:  cacheCheckoutEvery,
		},
	}, nil

}

func OpenConfig(path string) (*viper.Viper, error) {
	env := os.Getenv("ENV")
	if env == "" {
		return nil, fmt.Errorf("missing ENV env")
	}

	if env == "production" {
		path = fmt.Sprintf("prod.%s", path)
	}

	viper.SetConfigFile(path)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return viper.GetViper(), nil
}
