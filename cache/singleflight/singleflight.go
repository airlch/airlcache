package singleflight

import (
	"sync"
)

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// call 代表正在进行中，或已经结束的请求。使用 sync.WaitGroup 锁避免重入。
// Group 是 singleflight 的主数据结构，管理不同 key 的请求(call)。

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		//如果有数据，说明有并发请求，等待其完成，一起返回即可
		c.wg.Wait()
		return c.val, c.err
	}

	var c = new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}

// Do 方法，接收 2 个参数，第一个参数是 key，第二个参数是一个函数 fn。
// Do 的作用就是，针对相同的 key，在fn函数返回前，无论 Do 被调用多少次，函数 fn 都只会被调用一次，
// 等待 fn 调用结束了，返回返回值或错误。
