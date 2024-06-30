package main

import (
	"sync"

	"github.com/kiritoxkiriko/comical-kv/lru"
)

// cache is a thread-safe cache that holds ByteView, a wrapper for lru.Cache
type cache struct {
	lock       sync.RWMutex
	lru        *lru.Cache[ByteView]
	cacheBytes int64
}

// add adds a key-value pair to the cache
func (c *cache) add(key string, value ByteView) {
	c.lock.Lock()
	defer c.lock.Unlock()
	// lazy init
	if c.lru == nil {
		c.lru = lru.New[ByteView](c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

// get returns a value from the cache
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	// if not init just return
	if c.lru == nil {
		return
	}
	return c.lru.Get(key)
}
