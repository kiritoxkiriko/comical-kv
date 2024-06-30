package comical_kv

import (
	"fmt"
	"log"
	"sync"
)

// Getter is an interface that gets a value for a key
type Getter interface {
	// Get returns the value for a key
	Get(key string) ([]byte, error)
}

// GetterFunc is a function that implements Getter
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group is a cache group
type Group struct {
	// name is the name of the group
	name string
	// getter is the Getter of the group
	getter Getter
	// cache is the cache of the group
	cache cache
}

var (
	// rwLock is a mutex to protect groups
	rwLock sync.RWMutex
	// groups is a map of group names to groups, used to register groups
	groups = map[string]*Group{}
)

// NewGroup creates a new Group with a given name, cache size, and Getter
func NewGroup(name string, cacheBytes int64, getter Getter) (*Group, error) {
	if getter == nil {
		return nil, fmt.Errorf("nil getter")
	}
	// add a group to groups, using rwLock to protect
	g := &Group{
		name:   name,
		getter: getter,
		cache:  cache{cacheBytes: cacheBytes},
	}
	// register group
	rwLock.Lock()
	defer rwLock.Unlock()
	groups[name] = g
	return g, nil
}

// GetGroup returns a group by name, and a boolean indicating if the group exists
func GetGroup(name string) (*Group, bool) {
	rwLock.RLock()
	defer rwLock.RUnlock()
	g, ok := groups[name]
	return g, ok
}

// Get returns a value for a key from a group
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	// get from cache
	if v, ok := g.cache.get(key); ok {
		log.Println("[Comical KV] hit cache")
		return v, nil
	}
	// get from getter
	return g.load(key)
}

// load loads a value for a key from a group
func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

// getLocally gets a value for a key from a group's getter
func (g *Group) getLocally(key string) (value ByteView, err error) {
	// get from local first
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value = ByteView{b: cloneBytes(bytes)}
	// populate cache
	g.populateCache(key, value)
	return
}

// populateCache adds a key-value pair to a group's cache
func (g *Group) populateCache(key string, value ByteView) {
	g.cache.add(key, value)
}
