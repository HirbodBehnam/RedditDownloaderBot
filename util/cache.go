package util

import (
	"sync"
	"time"
)

// TimedCache is a KV cache which it's elements are deleted after a specified time
type TimedCache[K comparable, V any] struct {
	// The cache itself
	cache map[K]timedCacheElement[V]
	// A mutex to sync stuff
	lock sync.Mutex
}

// timedCacheElement is each element in the map of
type timedCacheElement[V any] struct {
	data      V
	addedTime time.Time
}

// NewTimedCache will create a timed cache of given type of key and value.
//
// ttl is the time which entries will live and cleanUpInterval is the time which old values will
// be purged from cache
func NewTimedCache[K comparable, V any](ttl, cleanUpInterval time.Duration) *TimedCache[K, V] {
	c := &TimedCache[K, V]{
		cache: make(map[K]timedCacheElement[V]),
	}
	go c.cleanUp(ttl, cleanUpInterval)
	return c
}

// cleanUp will delete old entries from cache
func (c *TimedCache[K, V]) cleanUp(ttl, cleanUpInterval time.Duration) {
	for {
		time.Sleep(cleanUpInterval)
		c.lock.Lock()
		for k, v := range c.cache {
			if time.Since(v.addedTime) > ttl {
				delete(c.cache, k)
			}
		}
		c.lock.Unlock()
	}
}

// Set will set a key and overwrite any old entries in cache
func (c *TimedCache[K, V]) Set(key K, value V) {
	c.lock.Lock()
	c.cache[key] = timedCacheElement[V]{
		data:      value,
		addedTime: time.Now(),
	}
	c.lock.Unlock()
}

// GetAndDelete will atomically get and element and delete it from cache
func (c *TimedCache[K, V]) GetAndDelete(key K) (V, bool) {
	c.lock.Lock()
	data, ok := c.cache[key]
	if ok {
		delete(c.cache, key)
	}
	c.lock.Unlock()
	return data.data, ok
}
