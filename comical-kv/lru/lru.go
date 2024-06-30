package lru

import (
	"container/list"
)

// Cache is an LRU cache. It is not safe for concurrent access
type Cache struct {
	// maxBytes is the maximum memory the cache can use
	maxBytes int64
	// nBytes is the current memory the cache is using
	nBytes int64
	// ll is the double linked list
	ll *list.List
	// cache is a map that maps a key to a list element
	cache map[string]*list.Element
	// OnEvicted is a callback function called when a key is evicted
	onEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	// Len returns the bytes that the value takes up
	Len() int
}

// New is the constructor of Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	// get elem from map
	if elem, ok := c.cache[key]; ok {
		// move elem to front
		c.ll.MoveToFront(elem)
		// get entry
		ent := elem.Value.(*entry)
		// return val
		return ent.value, true
	}
	return
}

func (c *Cache) RemoveOldest() {
	// get oldest
	elem := c.ll.Back()
	// if ll not empty
	if elem != nil {
		// remove from ll
		c.ll.Remove(elem)
		ent := elem.Value.(*entry)
		// remove from map
		delete(c.cache, ent.key)
		// deduct byte
		c.nBytes -= int64(len(ent.key) + ent.value.Len())
		if c.onEvicted != nil {
			c.onEvicted(ent.key, ent.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if elem, ok := c.cache[key]; ok {
		// move to front of ll
		c.ll.MoveToFront(elem)
		ent := elem.Value.(*entry)
		// add byte diff
		c.nBytes += int64(value.Len() - ent.value.Len())
		// update val
		ent.value = value
	} else {
		// set elem
		elem = c.ll.PushFront(&entry{
			key:   key,
			value: value,
		})
		// add byte
		c.nBytes += int64(value.Len() + len(key))
	}
	// remove overflow
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}
