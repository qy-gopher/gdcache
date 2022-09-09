package singleflight

import "sync"

// call 正在进行或已经结束的请求
type call struct {
	wg  sync.WaitGroup
	val any
	err error
}

// Group 管理不同key的call
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do 相同的key只能调用一次fn
func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()

	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)

	c.wg.Add(1)
	g.m[key] = c // 添加映射，表明key已有对应的请求在处理
	g.mu.Unlock()
	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
