package strategy

import (
	"container/list"
)

var _ Strategy = (*LRU)(nil)

// LRU lru 的结构实现
type LRU struct {
	capBytes  int64
	lenBytes  int64
	ll        *list.List
	dataMap   map[string]*list.Element
	onDeleted func(key string, value Value)
}

// lruEntry lru 中 list 的数据部分
type lruEntry struct {
	key   string
	value Value
}

// NewLRU 创建一个 lru
//
// cap：lru 缓存的最大比特数，0 表示无限制
// on: 删除数据时的回调函数
func NewLRU(cap int64, on func(key string, value Value)) *LRU {
	return &LRU{
		capBytes:  cap, // 0 means no limit
		ll:        list.New(),
		dataMap:   make(map[string]*list.Element),
		onDeleted: on,
	}
}

// Get 从 lru 中取数据
func (c *LRU) Get(key string) (Value, bool) {
	if ele, ok := c.dataMap[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*lruEntry)
		return kv.value, true
	}
	return nil, false
}

// DelOldest 淘汰 lru 中最久没被访问的数据
func (c *LRU) delOldest() {
	ele := c.ll.Back()
	if ele != nil {
		kv := ele.Value.(*lruEntry)
		delete(c.dataMap, kv.key)
		c.ll.Remove(ele)
		c.lenBytes -= int64(kv.value.Len()) + int64(len(kv.key))
		if c.onDeleted != nil {
			c.onDeleted(kv.key, kv.value)
		}
	}
}

// Add 添加数据到 lru
func (c *LRU) Add(key string, value Value) {
	if ele, ok := c.dataMap[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*lruEntry)
		c.lenBytes += int64(value.Len() - kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&lruEntry{key, value})
		c.dataMap[key] = ele
		c.lenBytes += int64(value.Len()) + int64(len(key))
	}

	for c.capBytes != 0 && c.lenBytes > c.capBytes {
		c.delOldest()
	}
}

// Len 返回 lru 中数据个数
func (c *LRU) Len() int {
	return c.ll.Len()
}
