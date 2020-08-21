package consistenthash

import (
	"strconv"
	"testing"
)

// func TestConsistenthash(t *testing.T) {
// 	var hashMap = New(3, nil)

// 	hashMap.Add("第一台", "第二台", "第三台")

// 	for i, v := range hashMap.hashMap {
// 		log.Printf("虚拟节点：%d   真实节点：%s", i, v)
// 	}

// 	t.Fatalf("haha")
// 	//var info = hashMap.Get("3423")
// }

func TestHashing(t *testing.T) {
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	// Given the above hash function, this will give replicas with "hashes":
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	hash.Add("6", "4", "2")

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

	// Adds 8, 18, 28
	hash.Add("8")

	// 27 should now map to 8.
	testCases["27"] = "8"

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

}
