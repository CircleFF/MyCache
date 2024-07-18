package strategy

import (
	"container/heap"
	"testing"
)

var _ Value = (*String)(nil)

func TestPriorityQueue(t *testing.T) {
	pq := priorityQueue{}
	heap.Init(&pq)
	heap.Push(&pq, &entry{
		key:   "k1",
		value: String("v1"),
		freq:  3,
	})
	if pq.Len() != 1 || pq[0].key != "k1" {
		t.Fatal("push failed")
	}

	heap.Push(&pq, &entry{
		key:   "k3",
		value: String("v3"),
		freq:  1,
	})
	if v := heap.Pop(&pq).(*entry); v.index != -1 || v.key != "k3" {
		t.Fatalf("sort and pop failed, v:%v", v)
	}

}
