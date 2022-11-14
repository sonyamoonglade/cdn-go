package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"animakuro/cdn/config"
	"animakuro/cdn/internal/cdn"
	"animakuro/cdn/internal/fs"
	"animakuro/cdn/internal/modules"
	cache "animakuro/cdn/pkg/cache/bucket"
	filecache "animakuro/cdn/pkg/cache/file"
	"animakuro/cdn/pkg/http_server"
	"animakuro/cdn/pkg/logging"
	"animakuro/cdn/pkg/middleware"
	"animakuro/cdn/pkg/mongodb"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/sonyamoonglade/dealer-go/v2"
)

func main() {

	log.Println("booting application...")

	debug, strict, logsPath, bucketsPath := parseFlags()
	fs.SetBucketsPath(bucketsPath)

	logger, err := logging.WithConfig(&logging.Config{
		Encoding: logging.JSON,
		Strict:   strict,
		LogsPath: logsPath,
		Debug:    debug,
	})

	if err != nil {
		log.Fatalf("could not get logger. %s", err.Error())
	}

	if err := godotenv.Load(".env"); err != nil {
		logger.Warnf("could not load .env variables %s", err.Error())
	}

	v, err := config.OpenConfig(config.BasePath)
	if err != nil {
		logger.Fatalf("could not open config: %s", err.Error())
	}

	cfg, err := config.GetAppConfig(v, debug)
	if err != nil {
		logger.Fatalf("could not get app config. %s", err.Error())
	}

	//Context for determining timeout to connect to database
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	mng, err := mongodb.New(ctx, cfg.MongoURI)
	if err != nil {
		logger.Fatalf("could not connect to database. %s", err.Error())
	}
	logger.Info("mongodb has connected")

	m := mux.NewRouter()
	srv := http_server.New(cfg.AppPort, cfg.AppHost, m)

	bucketCache := cache.NewBucketCache()
	fileCache := filecache.NewFileCache(logger, cfg.FileCacheConfig)
	middlewares := middleware.NewMiddlewares(logger, bucketCache)

	jobDealer := dealer.New(logger, cfg.MaxWorkers)
	jobDealer.WithStrategy(dealer.WorkerPool) // see dealer.WorkerPool impl.

	repo := cdn.NewRepository(logger, cfg.DBName, mng.Client())
	service := cdn.NewService(logger, repo, bucketCache, fileCache, cfg.Domain, jobDealer)
	handler := cdn.NewHandler(logger, m, service, cfg.MemoryConfig, middlewares, bucketCache, fileCache)

	err = service.InitBuckets(ctx)
	if err != nil {
		logger.Warnf("could not init buckets: %s", err.Error())
	}

	//Starts several goroutines
	err = fileCache.Start(debug)
	if err != nil {
		logger.Fatalf("could not start fileCache: %s", err.Error())
	}

	handler.InitRoutes()

	modules.Init()

	//Init worker pool and job pool
	jobDealer.Start(debug)

	//Graceful shutdown
	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server could not start listening. %s", err.Error())
		}
	}()
	logger.Infof("server is listening on %s:%s", cfg.AppHost, cfg.AppPort)

	<-shutdown
	logger.Debug("shutting down gracefully...")

	//Context for graceful shutdown
	gctx, gcancel := context.WithTimeout(context.Background(), time.Second*5)
	defer gcancel()

	if err := srv.Shutdown(gctx); err != nil {
		//No fatal error to make clean up
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

func parseFlags() (bool, bool, string, string) {

	debug := flag.Bool("debug", true, "determines whether logs are written to stdout or file")
	strict := flag.Bool("strict-log", false, "determines if logger shouldn't log any info/debug logs")
	logsPath := flag.String("logs-path", "", "determines where log file is")
	bucketsPath := flag.String("buckets-path", "", "determines where /buckets folder is")
	flag.Parse()

	//Critical for app if not specified
	if *bucketsPath == "" {
		panic("buckets path is not provided")
	}

	return *debug, *strict, *logsPath, *bucketsPath
}
