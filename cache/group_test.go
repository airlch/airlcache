package cache

import (
	"errors"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	var expect = []byte("kkk")
	if v, _ := f.Get("kkk"); !reflect.DeepEqual(v, expect) {
		t.Errorf("错了错了")
	}
}

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	var achche = NewGroup("course", 1000, GetterFunc(func(key string) ([]byte, error) {
		log.Println("now search:", key)
		if v, ok := db[key]; ok {
			if _, okload := loadCounts[key]; !okload {
				loadCounts[key] = 0
			} else {
				loadCounts[key]++
			}

			return []byte(v), nil
		}

		return []byte{}, errors.New("未找到缓存")
	}))

	for key, val := range db {
		if v, err := achche.Get(key); err != nil || v.String() != val {
			t.Fatal("failed to get value")
		}

		if i, ok := loadCounts[key]; !ok || i > 1 {
			t.Fatal("failed to get value")
		}
	}

	if _, err := achche.Get("aaaa"); err == nil {
		t.Fatal("error")
	}

}
