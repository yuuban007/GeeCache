package lru

import (
	"reflect"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}
func TestAdd(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("key1", String("12345"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "12345" {
		t.Fatalf("cache hit key1=12345 failed")
	}
	if lru.Len() != 1 {
		t.Fatalf("cache length after Add is incorrect, expected 1, got %d", lru.Len())
	}

	lru.Add("key2", String("67890"))
	if v, ok := lru.Get("key2"); !ok || string(v.(String)) != "67890" {
		t.Fatalf("cache hit key2=67890 failed")
	}
	if lru.Len() != 2 {
		t.Fatalf("cache length after Add is incorrect, expected 2, got %d", lru.Len())
	}
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("key1", String("12345"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "12345" {
		t.Fatalf("cache hit key1=12345 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("key1", String("12345"))
	lru.Add("key2", String("67890"))
	lru.RemoveOldest()

	if _, ok := lru.Get("key1"); ok {
		t.Fatalf("cache miss key1 failed")
	}
	if v, ok := lru.Get("key2"); !ok || string(v.(String)) != "67890" {
		t.Fatalf("cache hit key2=67890 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	// Test callback function
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}

	lru := New(int64(10), callback)
	lru.Add("Key1", String("123"))
	lru.Add("Key1", String("12"))
	lru.Add("Key2", String("a"))
	lru.Add("Key3", String("a"))
	lru.Add("Key4", String("a"))

	expect := []string{"Key1", "Key2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("call onEvicted failed, expect keys equals to %s,Got %s", expect, keys)
	}
}
