package lru

import "container/list"

// Cache is a LRU cache, It is not safe for concurrent access.
type Cache struct {
	maxBytes int64
	nbytes   int64
	ll       *list.List
	cache    map[string]*list.Element
	// OnEvicted is an optional function that is executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

// entry represents a key-value pair in the LRU cache.
type entry struct {
	key   string
	value Value
}

// Value is an interface that represents the value stored in the cache.
// Implementations of Value should provide a Len() method to count the number of bytes it takes.
type Value interface {
	Len() int
}

// New creates a new LRU cache with the specified maximum number of bytes and an optional eviction callback function.
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     map[string]*list.Element{},
		OnEvicted: onEvicted,
	}
}

// Get looks up a key's value in the cache.
// If the key exists, the corresponding entry is moved to the front of the cache (most recently used).
// Returns the value and true if the key exists, or nil and false otherwise.
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest removes the oldest entry from the cache.
// The oldest entry is the one at the back of the cache (least recently used).
// If an eviction callback function is specified, it is executed with the key and value of the removed entry.
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele = c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
