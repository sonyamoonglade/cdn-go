package cache

import (
	"sync"

	"animakuro/cdn/internal/entities"
	"animakuro/cdn/pkg/cdn_errors"
)

type BucketCache struct {
	cache map[string]*entities.Bucket
	mu    *sync.RWMutex
}

func NewBucketCache() *BucketCache {
	return &BucketCache{
		cache: make(map[string]*entities.Bucket),
		mu:    new(sync.RWMutex),
	}
}

func (bc *BucketCache) Get(bucketName string) (*entities.Bucket, error) {
	b, ok := bc.cache[bucketName]
	if ok == false {
		return nil, cdn_errors.ErrBucketNotFound
	}
	return b, nil
}

func (bc *BucketCache) Add(b *entities.Bucket) {
	bc.mu.Lock()
	bc.cache[b.Name] = b
	bc.mu.Unlock()
}
