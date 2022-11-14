package logging

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Encoding string

const (
	JSON    Encoding = "json"
	Console          = "console"
)

type Config struct {
	Encoding Encoding
	Strict   bool
	Debug    bool
	LogsPath string
}

func WithConfig(cfg *Config) (*zap.SugaredLogger, error) {
	builder := zap.NewProductionConfig()

	builder.Encoding = string(cfg.Encoding)
	builder.Development = cfg.Debug

	builder.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	if cfg.Strict {
		builder.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	}

	//Prod
	if !cfg.Debug {
		_, err := os.Stat(cfg.LogsPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				panic(fmt.Sprintf("log file at %s does not exist", cfg.LogsPath))
			}
			return nil, err
		}

		builder.OutputPaths = []string{cfg.LogsPath}
	}

	logger, err := builder.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}
