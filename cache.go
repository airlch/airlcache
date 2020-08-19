package airlcache

import (
	"airlsubject/airlcache/lru"
	"errors"
	"sync"
)

type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) Add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(100, nil)
	}

	c.lru.Add(key, value)
}

func (c *cache) Get(key string) (byteview ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return nil, false
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}

	return nil, false
}
