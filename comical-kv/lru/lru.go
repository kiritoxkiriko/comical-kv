package lru

import (
	"container/list"
	"go/types"
)

// Cache is an LRU cache. It is not safe for concurrent access
type Cache struct {
	// maxBytes is the maximum memory the cache can use
	maxBytes int64
	// nBytes is the current memory the cache is using
	nBytes int64
	// ll is the double linked list
	ll list.List
	// cache is a map that maps a key to a list element
	cache map[string]*list.Element
	// OnEvicted is a callback function that is called when a key is evicted
	onEvicted func(key string, value types.Type)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}
