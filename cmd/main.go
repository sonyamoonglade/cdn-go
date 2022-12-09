package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"animakuro/cdn/config"
	"animakuro/cdn/internal/cdn"
	"animakuro/cdn/internal/fs"
	"animakuro/cdn/internal/modules"
	bucketcache "animakuro/cdn/pkg/cache/bucket"
	filecache "animakuro/cdn/pkg/cache/file"
	"animakuro/cdn/pkg/dealer"
	"animakuro/cdn/pkg/http"
	"animakuro/cdn/pkg/logging"
	"animakuro/cdn/pkg/middleware"
	"animakuro/cdn/pkg/mongodb"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {

	log.Println("booting application...")

	debug, strictLogging, logsPath, bucketsPath, configPath := parseFlags()
	fs.SetBucketsPath(bucketsPath)

	logger, err := logging.WithConfig(&logging.Config{
		Encoding: logging.JSON,
		Strict:   strictLogging,
		LogsPath: logsPath,
		Debug:    debug,
	})
	if err != nil {
		log.Fatalf("could not get logger. %s", err.Error())
	}
	defer logger.Sync()

	// Try to get .env variables on development only mode
	if err := godotenv.Load(".env"); err != nil {
		logger.Warnf("could not load .env variables %s", err.Error())
	}

	cfg, err := config.GetAppConfig(configPath, debug)
	if err != nil {
		logger.Fatalf("could not get app config. %s", err.Error())
	}
	logger.Info(cfg)

	//Context for determining timeout to connect to database
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	mng, err := mongodb.New(ctx, cfg.MongoURI)
	if err != nil {
		logger.Fatalf("could not connect to database. %s", err.Error())
	}
	logger.Info("mongodb has connected")

	m := mux.NewRouter()
	srv := http.NewServer(cfg.AppPort, cfg.AppHost, m)

	bucketCache := bucketcache.NewBucketCache()
	fileCache := filecache.NewFileCache(logger, cfg.FileCacheConfig)
	middlewares := middleware.NewMiddlewares(logger, bucketCache)
	moduleController := modules.NewController()

	// Worker pool for IO operations
	jobDealer := dealer.New(logger, cfg.MaxWorkers)
	jobDealer.WithStrategy(dealer.WorkerPool)

	repo := cdn.NewRepository(logger, cfg.DBName, mng.Client())
	service := cdn.NewService(logger, repo, bucketCache, fileCache, cfg.Domain, jobDealer)
	handler := cdn.NewHandler(&cdn.HandlerDeps{
		Logger:           logger,
		Mux:              m,
		Middlewares:      middlewares,
		Service:          service,
		ModuleController: moduleController,
		BucketCache:      bucketCache,
		FileCache:        fileCache,
		MemConfig:        cfg.MemoryConfig,
	})

	err = service.InitBuckets(ctx)
	if err != nil {
		logger.Warnf("could not init buckets: %s", err.Error())
	}

	err = fileCache.Start(debug)
	if err != nil {
		logger.Fatalf("could not start fileCache: %s", err.Error())
	}

	handler.InitRoutes()

	// Init worker pool and job pool
	jobDealer.Start(debug)

	// Graceful shutdown
	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("server could not start listening. %s", err.Error())
		}
	}()
	logger.Infof("server is listening on %s:%s", cfg.AppHost, cfg.AppPort)

	<-shutdown
	logger.Debug("shutting down gracefully...")

	// Context for graceful shutdown
	gctx, gcancel := context.WithTimeout(context.Background(), time.Second*5)
	defer gcancel()

	if err := srv.Shutdown(gctx); err != nil {
		// No fatal error to make clean up
		logger.Errorf("server could not shutdown gracefully. %s", err.Error())
	}
	logger.Debug("server has shutdown")

	if err := mng.CloseConnection(gctx); err != nil {
		logger.Errorf("mongo could not close connection. %s", err.Error())
	}
	logger.Debug("mongo connection has closed")

	fileCache.Stop()
	logger.Debugf("fileCache has stopped")

	jobDealer.Stop()
	logger.Debugf("jobDealer has stopped")
}

func parseFlags() (bool, bool, string, string, string) {

	debug := flag.Bool("debug", true, "determines whether logs are written to stdout or file")
	strictLogging := flag.Bool("strict-log", false, "determines if logger shouldn't log any info/debug logs")
	logsPath := flag.String("logs-path", "", "determines where log file is")
	bucketsPath := flag.String("buckets-path", "", "determines where /buckets folder is")
	configPath := flag.String("config-path", "", "determines where config file is")

	flag.Parse()

	// Critical for app if not specified
	if *bucketsPath == "" {
		panic("buckets path is not provided")
	}

	// Critical for app if not specified
	if *configPath == "" {
		panic("config path is not provided")
	}

	// Naked return, see return variable names
	return *debug, *strictLogging, *logsPath, *bucketsPath, *configPath
}
