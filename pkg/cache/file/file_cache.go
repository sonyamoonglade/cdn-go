package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/gokyle/filecache"
	"go.uber.org/zap"
)

type Config struct {
	MaxCacheSize   int64
	MaxCacheItems  int
	CacheTTL       int
	CacheThreshold int
	FlushEvery     int
	CheckoutEvery  int
}

type FileCache struct {
	cache        *filecache.FileCache
	hitThreshold int
	hits         map[string]int
	mu           *sync.RWMutex
	logger       *zap.SugaredLogger
	isFlushing   bool
	ticker       *time.Ticker
	wg           *sync.WaitGroup
	shutdown     chan interface{}
}

func NewFileCache(logger *zap.SugaredLogger, cfg *Config) *FileCache {

	c := filecache.NewDefaultCache()
	c.MaxItems = cfg.MaxCacheItems
	c.Every = cfg.CheckoutEvery
	c.MaxSize = cfg.MaxCacheSize * filecache.Megabyte
	c.ExpireItem = cfg.CacheTTL

	return &FileCache{
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

//Hit increments hits for certain item by specific path
func (fc *FileCache) Hit(path string) {
	var beats bool
	isCached := fc.cache.InCache(path)

	curr, ok := fc.hits[path]
	if !ok {
		fc.mu.Lock()
		fc.hits[path] = 1
		fc.mu.Unlock()
	}

	if !isCached && curr < fc.hitThreshold {
		fc.mu.Lock()
		fc.hits[path] = curr + 1
		fc.mu.Unlock()
		//If just incremented value is equal threshold - set beats
		if curr+1 == fc.hitThreshold {
			beats = true
		}
	}

	//Either beats or counter is already greater
	if !isCached && (beats || curr >= fc.hitThreshold) {
		fc.cache.Cache(path)
		delete(fc.hits, path)
	}

	return
}

func (fc *FileCache) flush() {
	fc.mu.Lock()
	for k := range fc.hits {
		delete(fc.hits, k)
	}
	fc.mu.Unlock()
}

func (fc *FileCache) flushing() {
	for {
		select {
		case <-fc.ticker.C:
			fc.logger.Debugf("flushing cache")
			fc.flush()
		case <-fc.shutdown:
			fc.ticker.Stop()
			fc.wg.Done()
			return
		}
	}
}

func (fc *FileCache) debug() {
	debugTicker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-debugTicker.C:
			mem := float64(fc.cache.FileSize()) / (float64(1024 * 1024))
			fc.logger.Debugf("cached items: %d cache size: %.4fMB", fc.cache.Size(), mem)
		case <-fc.shutdown:
			fc.ticker.Stop()
			fc.wg.Done()
			return
		}

	}
}

func (fc *FileCache) Start(debug bool) error {
	if fc.isFlushing {
		return nil
	}

	err := fc.cache.Start()
	if err != nil {
		return fmt.Errorf("FileCache.fc.cache.Start")
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

func (fc *FileCache) Lookup(path string) ([]byte, bool) {

	isCached := fc.cache.InCache(path)
	if !isCached {
		return nil, false
	}

	return fc.cache.GetItem(path)
}

func (fc *FileCache) Stop() {
	//Clear hits map
	fc.mu.Lock()
	fc.hits = nil
	fc.mu.Unlock()

	close(fc.shutdown)
	fc.wg.Wait()

	//Clear file cache and it's underlying stuff
	fc.cache.Stop()
}
