package lru

import (
	"container/list"
)

type Cache struct {
	//允许使用最大内存
	maxBytes int64
	//当前内存大小
	nBytes int64

	//双向链表(double linked list)
	ll *list.List
	//缓存字典
	cache map[string]*list.Element
	// OnEvicted 是某条记录被移除时的回调函数，可以为 nil
	OnEvicted func(key string, value interface{})
}

//具体缓存kv
type entity struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

//初始化
func New(maxbytes int64, onEvicted func(string, interface{})) *Cache {
	return &Cache{
		maxBytes:  maxbytes,
		nBytes:    0,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

//获取缓存
func (c *Cache) Get(key string) (interface{}, bool) {
	if ele, ok := c.cache[key]; ok {
		//移到队尾，     lru，最近最少使用      新增获取放队尾，移除移队首
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entity)
		return kv.value, true
	}
	return nil, false
}

//移除
func (c *Cache) Remove(key string) {
	if ele, ok := c.cache[key]; ok {
		c.ll.Remove(ele)
		var kv = ele.Value.(*entity)
		delete(c.cache, kv.key)
		c.nBytes -= int64(kv.value.Len())
	}
}

//移除最新不常使用的
func (c *Cache) removeOldest() {
	var ele = c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		var kv = ele.Value.(*entity)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		delete(c.cache, kv.key)
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

//新增修改
func (c *Cache) Add(key string, val Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entity)
		c.nBytes += int64(val.Len()) - int64(kv.value.Len())
		kv.value = val
	} else {
		ele := c.ll.PushFront(&entity{key, val})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(val.Len())
	}

	for c.maxBytes > 0 && c.nBytes > c.maxBytes {
		c.removeOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
