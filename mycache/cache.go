package mycache

import (
	"MyCache/mycache/strategy"
	"sync"
)

// cache lru 上层的封装
type cache struct {
	mu         sync.Mutex
	strategy   strategy.Strategy
	cacheBytes int64
}

func newCache(strategyName string, cap int64) *cache {
	on := func(key string, value strategy.Value) {}
	return &cache{
		mu:         sync.Mutex{},
		strategy:   strategy.New(strategyName, cap, on),
		cacheBytes: cap,
	}
}

// set 添加缓存数据
func (c *cache) set(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.strategy.Add(key, value)
}

// get 获取缓存数据
func (c *cache) get(key string) (ByteView, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.strategy.Get(key); ok {
		return v.(ByteView), ok
	}

	return ByteView{}, false
}
