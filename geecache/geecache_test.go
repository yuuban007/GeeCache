package geecache

import (
	"fmt"
	"log"
	"testing"
)

var db = map[string]string{
	"key1": "value1",
	"key2": "value2",
	"key3": "value3",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	f := GetterFunc(func(key string) ([]byte, error) {
		log.Println("[Whatever DB] search key", key)
		if v, ok := db[key]; ok {
			if _, ok := loadCounts[key]; !ok {
				loadCounts[key] = 0
			}
			loadCounts[key] += 1
			return []byte(v), nil
		}
		return []byte{}, fmt.Errorf("%s is not exists", key)
	})
	gee := NewGroup("testGroup", 2<<10, f)
	for k, v := range db {
		if view, err := gee.Get(k); err != nil || view.String() != v {
			t.Fatalf("Failed to get value of %s", k)
		} //load from callback function
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknown should be empty, but %s got", view)
	}
}
