package strategy

import (
	"container/heap"
)

var _ Strategy = (*LFU)(nil)

type LFU struct {
	capBytes  int64
	lenBytes  int64
	pq        *priorityQueue
	dataMap   map[string]*entry
	onDeleted func(key string, value Value)
}

func NewLFU(cap int64, on func(key string, value Value)) *LFU {
	priorityQueue := make(priorityQueue, 0)
	return &LFU{
		capBytes:  cap,
		pq:        &priorityQueue,
		dataMap:   make(map[string]*entry),
		onDeleted: on,
	}
}

func (L *LFU) Get(key string) (Value, bool) {
	if e, ok := L.dataMap[key]; ok {
		e.freq++
		heap.Fix(L.pq, e.index)
		return e.value, true
	}
	return nil, false
}

func (L *LFU) Add(key string, value Value) {
	if e, ok := L.dataMap[key]; ok {
		e.freq++
		L.lenBytes += int64(value.Len() - e.value.Len())
		e.value = value
		heap.Fix(L.pq, e.index)
	} else {
		e = &entry{
			key:   key,
			value: value,
			freq:  1,
		}
		L.lenBytes += int64(len(key) + value.Len())
		heap.Push(L.pq, e)
		L.dataMap[key] = e
	}

	for L.capBytes != 0 && L.lenBytes > L.capBytes {
		L.delOldest()
	}
}

func (L *LFU) Len() int {
	return L.pq.Len()
}

func (L *LFU) delOldest() {
	e := heap.Pop(L.pq).(*entry)
	delete(L.dataMap, e.key)
	L.lenBytes -= int64(len(e.key)) + int64(e.value.Len())
	if L.onDeleted != nil {
		L.onDeleted(e.key, e.value)
	}
}
