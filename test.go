package main

// import (
// 	"container/list"
// 	"fmt"
// )

// func main() {
// 	//a := 0
// 	//fmt.Scan(&a)
// 	//fmt.Printf("%d\n", a)
// 	fmt.Printf("Hello World!\n")
// 	ch := make(chan int)
// 	channel := make(chan int, 1)
// 	c := &LRUCacheConstructor(2)
// 	c.get(1)
// 	c.put(1, 2)
// }

// type LRUCache struct {
// 	capacity int
// 	cache    map[int]*list.Element
// 	ll       *list.List
// }

// type entry struct {
// 	key   int
// 	value int
// }

// func LRUCacheConstructor(capacity int) LRUCache {
// 	return LRUCache{capacity, make(map[int]*list.Element), new(list.List)}
// }

// func (c *LRUCache) Len() int {
// 	return len(c.cache) // or len(c.ll)
// }

// // 移除最不经常使用的
// func (c *LRUCache) Remove() {
// 	ele := c.ll.Back()
// 	if ele == nil {
// 		return
// 	}
// 	c.ll.Remove(ele) // 删除队尾元素
// 	kv := ele.Value()
// 	delete(c.cache, kv.key) // 删除map中的key
// }

// func (c *LRUCache) get(key int) (value int) {
// 	if ele, ok := c.cache[key]; ok {
// 		c.ll.MoveToFront(ele) // 移动到队头
// 		return ele.value().value
// 	}
// 	return -1
// }

// func (c *LRUCache) put(key int, value int) {
// 	// 存在key
// 	if ele, ok := c.cache[key]; ok {
// 		c.ll.MoveToFront(ele) // 移动到队头
// 		// 更新value
// 		ele.Value() = &entry{key, value}
// 	}
// 	// 不存在key
// 	ele := &entry{key, value}
// 	if c.Len() >= c.capacity {
// 		c.Remove()
// 	}
// 	// 插入到队头
// 	c.ll.PushFront(ele)
// 	c.cache[key] = value
// }
