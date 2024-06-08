package consistenthash

import (
	"strconv"
	"testing"
)

func TestGet(t *testing.T) {

	hashFunc := Hash(func(data []byte) uint32 {
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	})

	hashRing := New(3, hashFunc)
	hashRing.Add("2", "4", "6")

	// Given the above hash function, this will give replicas with "hashes":
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"25": "6",
		// 27
		"27": "2",
	}

	for k, v := range testCases {
		if hashRing.Get(k) != v {
			t.Errorf("Asking for %s,should have yielded %s,Get %s", k, v, hashRing.Get(k))
		}
	}

	// Add 8 18 28
	hashRing.Add("8")

	testCases["27"] = "8"
	for k, v := range testCases {
		if hashRing.Get(k) != v {
			t.Errorf("Asking for %s,should have yielded %s,Get %s", k, v, hashRing.Get(k))
		}
	}

}
