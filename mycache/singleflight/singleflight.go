package singleflight

import "sync"

// call 一次请求调用
type call struct {
	wg  *sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do 瞬时间内多次请求同一个 key，只进行一次 http，其他被阻塞的进程共享结果
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := &call{
		wg: &sync.WaitGroup{},
	}
	g.m[key] = c
	c.wg.Add(1)
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	// 加锁删除，防止后来的进程在 c.wg.wait() 时返回空数据异常，避免并发错误
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
