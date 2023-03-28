package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/gokyle/filecache"
	"go.uber.org/zap"
)

type FileCache interface {
	Increment(path string)
	Lookup(path string) ([]byte, bool)
	Stop()
}

type Config struct {
	MaxCacheSize   int64
	MaxCacheItems  int
	CacheTTL       int
	CacheThreshold int
	FlushEvery     int
	CheckoutEvery  int
}

type fileCache struct {
	cache        *filecache.FileCache
	hitThreshold int
	hits         map[string]int
	logger       *zap.SugaredLogger
	ticker       *time.Ticker
	mu           *sync.RWMutex
	wg           *sync.WaitGroup
	isFlushing   bool
	shutdown     chan interface{}
}

func NewFileCache(logger *zap.SugaredLogger, cfg *Config) *fileCache {

	c := filecache.NewDefaultCache()
	c.MaxItems = cfg.MaxCacheItems
	c.Every = cfg.CheckoutEvery
	c.MaxSize = cfg.MaxCacheSize * filecache.Megabyte
	c.ExpireItem = cfg.CacheTTL

	return &fileCache{
		cache:        c,
		hitThreshold: cfg.CacheThreshold,
		hits:         make(map[string]int),
		wg:           new(sync.WaitGroup),
		mu:           new(sync.RWMutex),
		logger:       logger,
		ticker:       time.NewTicker(time.Second * time.Duration(cfg.FlushEvery)),
		isFlushing:   false,
		shutdown:     make(chan interface{}),
	}
}

// Increment increments hits for certain file by specific path
func (fc *fileCache) Increment(path string) {
	var beats bool
	isCached := fc.cache.InCache(path)
	curr, ok := fc.hits[path]
	if !ok {
		fc.mu.Lock()
		fc.hits[path] = 1
		// TODO: inc metric gauge
		fc.mu.Unlock()
	}

	if !isCached && curr < fc.hitThreshold {
		fc.mu.Lock()
		// TODO: inc metric gauge
		fc.hits[path] = curr + 1
		fc.mu.Unlock()
		//If just incremented value is equal threshold - set beats
		if curr+1 == fc.hitThreshold {
			beats = true
		}
	}

	// Cache file if its not cached and current file's hits has just beaten the threshold or curr is greater
	if !isCached && (beats || curr >= fc.hitThreshold) {
		fc.cache.Cache(path)
		// TODO: Decrement metric gauge
		delete(fc.hits, path)
	}

	return
}

func (fc *fileCache) Start(debug bool) error {
	if fc.isFlushing {
		return nil
	}

	err := fc.cache.Start()
	if err != nil {
		return fmt.Errorf("fileCache.fc.cache.Start")
	}

	fc.isFlushing = true

	fc.wg.Add(1)
	go fc.flushing()

	if debug {
		fc.wg.Add(1)
		go fc.debug()
	}

	return nil
}

func (fc *fileCache) Lookup(path string) ([]byte, bool) {

	isCached := fc.cache.InCache(path)
	if !isCached {
		return nil, false
	}

	return fc.cache.GetItem(path)
}

func (fc *fileCache) Stop() {
	// Clear hits map
	fc.mu.Lock()
	fc.hits = nil
	fc.mu.Unlock()

	close(fc.shutdown)
	fc.wg.Wait()

	// Clear file cache and it's underlying stuff
	fc.cache.Stop()
}

func (fc *fileCache) flush() {
	fc.mu.Lock()
	// TODO: Decrement metric gauge to 0
	for k := range fc.hits {
		delete(fc.hits, k)
	}
	fc.mu.Unlock()
}

func (fc *fileCache) flushing() {
	for {
		select {
		case <-fc.ticker.C:
			fc.flush()
		case <-fc.shutdown:
			fc.ticker.Stop()
			fc.wg.Done()
			return
		}
	}
}

func (fc *fileCache) debug() {
	debugTicker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-debugTicker.C:
			// TODO: prometheus metric
		case <-fc.shutdown:
			fc.ticker.Stop()
			fc.wg.Done()
			return
		}

	}
}
