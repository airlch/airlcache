package lru

import (
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func testGet(t testing.T) {
	var lru = New(int64(0), nil)
	lru.Add("key1", String("111"))

	if v, ok := lru.Get("key1"); !ok || v.(string) != "1234" {
		t.Fatalf("cache hit key1=%s failed", "1234")
	}

	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveoldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(cap), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Get(k1)
	lru.Add(k3, String(v3))

	if _, ok := lru.Get("key2"); ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}
