package geecache

import (
	"fmt"
	pb "geecache/geecachepb"
	"geecache/singleflight"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
	// use singleflight.Group to make sure
	// that each key is fetched once at the same
	loader *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup creates a new cache group with the specified name, cache size, and getter function.
// It panics if the getter function is nil.
// The cache group is stored in the global groups map and returned.
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}

	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup returns the group with the specified name.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// RegisterPeers registers the PeerPicker for the Group.
// It panics if RegisterPeers is called more than once.
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeers called more than once")
	}
	g.peers = peers
}

// Get retrieves the value for the given key from the cache.
// If the key is empty, it returns an empty ByteView and an error indicating that the key is required.
// If the value is found in the cache, it returns the value (ByteView) and nil error.
// If the value is not found in the cache, it calls the load method to load the value and returns it.
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	return g.load(key)
}

// load loads the value for the given key from the cache.
// If the value is not found in the cache, it tries to retrieve it from the peers.
// If the peers are available and the value is found, it is stored in the cache and returned.
// If the peers are not available or the value is not found, it tries to retrieve it locally.
// If the value is found locally, it is stored in the cache and returned.
// All the fetch are done through the loader function to make sure that each key is fetched once at the same time.
func (g *Group) load(key string) (value ByteView, err error) {

	result, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return result.(ByteView), nil
	}
	return
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: res.Value}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
