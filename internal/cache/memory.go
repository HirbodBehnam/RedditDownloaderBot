package cache

import (
	"sync"
	"time"
)

var _ Interface = &MemoryCache{}

// singleMemoryCache is a KV cache which it's elements are deleted when cleanUp is called.
// Data is stored in ram
type singleMemoryCache[K comparable, V any] struct {
	// The cache itself
	cache map[K]memoryCacheElement[V]
	// A mutex to sync stuff
	lock sync.Mutex
}

// memoryCacheElement is each element in the map of in memory cache
type memoryCacheElement[V any] struct {
	data      V
	addedTime time.Time
}

// cleanUp will delete old entries from cache
func (c *singleMemoryCache[K, V]) cleanUp(ttl time.Duration) {
	c.lock.Lock()
	for k, v := range c.cache {
		if time.Since(v.addedTime) > ttl {
			delete(c.cache, k)
		}
	}
	c.lock.Unlock()
}

// set will set a key and overwrite any old entries in cache
func (c *singleMemoryCache[K, V]) set(key K, value V) {
	c.lock.Lock()
	c.cache[key] = memoryCacheElement[V]{
		data:      value,
		addedTime: time.Now(),
	}
	c.lock.Unlock()
}

// getAndDelete will atomically get and element and delete it from cache
func (c *singleMemoryCache[K, V]) getAndDelete(key K) (V, bool) {
	c.lock.Lock()
	data, ok := c.cache[key]
	if ok {
		delete(c.cache, key)
	}
	c.lock.Unlock()
	return data.data, ok
}

// MemoryCache is an in memory cache to handle the callback data
type MemoryCache struct {
	mediaCache singleMemoryCache[string, CallbackDataCached]
	albumCache singleMemoryCache[string, CallbackAlbumCached]
	// Close this channel to stop the cleanup
	cleanUpDoneChannel chan struct{}
}

// NewMemoryCache will create a timed cache of given type of key and value.
//
// ttl is the time which entries will live and cleanUpInterval is the time which old values will
// be purged from cache
func NewMemoryCache(ttl, cleanUpInterval time.Duration) *MemoryCache {
	c := &MemoryCache{
		mediaCache: singleMemoryCache[string, CallbackDataCached]{
			cache: make(map[string]memoryCacheElement[CallbackDataCached]),
		},
		albumCache: singleMemoryCache[string, CallbackAlbumCached]{
			cache: make(map[string]memoryCacheElement[CallbackAlbumCached]),
		},
		cleanUpDoneChannel: make(chan struct{}),
	}
	go c.cleanUp(ttl, cleanUpInterval)
	return c
}

// cleanUp must be executed in another goroutine. It blocks and waits either for
// cleanUpInterval seconds or MemoryCache.cleanUpDoneChannel is closed
func (c *MemoryCache) cleanUp(ttl, cleanUpInterval time.Duration) {
	cleanUpWait := time.NewTicker(cleanUpInterval)
	for {
		select {
		case <-cleanUpWait.C:
			c.albumCache.cleanUp(ttl)
			c.mediaCache.cleanUp(ttl)
		case <-c.cleanUpDoneChannel:
			cleanUpWait.Stop()
			return
		}
	}
}

func (c *MemoryCache) SetMediaCache(key string, value CallbackDataCached) error {
	c.mediaCache.set(key, value)
	return nil
}

func (c *MemoryCache) GetAndDeleteMediaCache(key string) (CallbackDataCached, error) {
	value, exists := c.mediaCache.getAndDelete(key)
	var err error
	if !exists {
		err = NotFoundErr
	}
	return value, err
}

func (c *MemoryCache) SetAlbumCache(key string, value CallbackAlbumCached) error {
	c.albumCache.set(key, value)
	return nil
}

func (c *MemoryCache) GetAndDeleteAlbumCache(key string) (CallbackAlbumCached, error) {
	value, exists := c.albumCache.getAndDelete(key)
	var err error
	if !exists {
		err = NotFoundErr
	}
	return value, err
}

// Close will cancel the clean-up goroutine
func (c *MemoryCache) Close() error {
	close(c.cleanUpDoneChannel)
	return nil
}
