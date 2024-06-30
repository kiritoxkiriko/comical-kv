package comical_kv

import (
	"sync"

	"github.com/kiritoxkiriko/comical-kv/lru"
)

// cache is a thread-safe cache that holds ByteView, a wrapper for lru.Cache
type cache struct {
	// lock is a mutex to protect the cache
	// NOTE: cannot use sync.RWMutex because get and add are not atomic
	lock       sync.Mutex
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
	c.lock.Lock()
	defer c.lock.Unlock()
	// if not init just return
	if c.lru == nil {
		return
	}
	return c.lru.Get(key)
}
