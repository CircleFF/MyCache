package strategy

import (
	"container/heap"
)

var _ heap.Interface = (*priorityQueue)(nil)

type entry struct {
	index int
	key   string
	value Value
	freq  int
}

type priorityQueue []*entry

func (p priorityQueue) Len() int {
	return len(p)
}

func (p priorityQueue) Less(i, j int) bool {
	return p[i].freq < p[j].freq
}

func (p priorityQueue) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
	p[i].index = i
	p[j].index = j

}

func (p *priorityQueue) Push(x any) {
	item := x.(*entry)
	item.index = len(*p)
	*p = append(*p, item)

}

func (p *priorityQueue) Pop() any {
	old := *p
	n := len(old)
	e := old[n-1]
	old[n-1].index = -1
	old[n-1] = nil
	*p = old[:n-1]
	return e
}
