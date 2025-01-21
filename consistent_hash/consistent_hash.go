package consistent_hash

import (
	"hash/crc32"
	"slices"
	"strconv"
)

// Hash is a function that takes a byte slice and returns an uint32.
type Hash func(data []byte) uint32

// Map is a hash ring of type Hash.
type Map struct {
	// hash function
	hash Hash
	// replicas virtual node scale
	replicas int
	// keys sorted keys
	keys []uint32
	// hashMap hash to key
	hashMap map[uint32]string
}

// New creates a new Map instance. If the hash function is nil, crc32.ChecksumIEEE is used.
func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  map[uint32]string{},
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some keys to the hash ring.
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := range m.replicas {
			// hash virtual key
			h := m.hash([]byte(strconv.Itoa(i) + key))
			// add hash to key ring
			m.keys = append(m.keys, h)
			// map hash to key
			m.hashMap[h] = key
		}
	}
	// sort key
	slices.Sort(m.keys)
}

// Get gets the closest item in the hash ring to the provided key.
func (m *Map) Get(key string) string {
	// if no key, return
	if len(m.keys) == 0 {
		return ""
	}

	h := m.hash([]byte(key))
	// use binary search found position
	idx, _ := slices.BinarySearch(m.keys, h)

	// get key, use mod in case overflow
	foundKey := m.keys[idx%len(m.keys)]
	return m.hashMap[foundKey]
}
