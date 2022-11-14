package middleware

import (
	cache "animakuro/cdn/pkg/cache/bucket"
	"animakuro/cdn/pkg/middleware/jwt"
	"go.uber.org/zap"
)

type Middlewares struct {
	JwtMiddleware *jwt.Middleware
}

func NewMiddlewares(logger *zap.SugaredLogger, bucketCache *cache.BucketCache) *Middlewares {
	jwtm := jwt.NewMiddleware(logger, bucketCache)

	return &Middlewares{JwtMiddleware: jwtm}
}
