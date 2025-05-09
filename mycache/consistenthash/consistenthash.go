package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// Map contains all hashed keys
type Map struct {
	hash     Hash
	replicas int
	// keys 存储虚拟节点哈希值
	keys []int // sorted
	// hashMap 键为虚拟节点哈希值，值为真实节点
	hashMap map[int]string
}

// New creates a map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  map[int]string{},
	}
	if fn == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some keys to hash.
// key mean physic machine or physic node
// TODO: replace keys to machine
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			// 计算虚拟节点的hash值
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get gets the closet item in hash to the provided key
func (m *Map) Get(key string) string {
	if len(key) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// Binary search for appropriate replica
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })
	// 二分法没有找到比虚拟节点hash值更大的索引式，返回的式keys的长度
	// 这时候取余等于0就回到了最开始的点
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
