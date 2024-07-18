package strategy

import (
	"reflect"
	"testing"
)

type String string

// Len 实现 Len() 方法，返回字符串占用的内存大小
func (s String) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	lru := NewLRU(int64(0), nil)
	lru.Add("k1", String("v1"))
	if v, ok := lru.Get("k1"); !ok || v.(String) != "v1" {
		t.Fatalf("mycache hit k1=v1 failed")
	}
	if _, ok := lru.Get("k2"); ok {
		t.Fatal("mycache miss k2 failed")
	}
}

func TestDelOldest(t *testing.T) {
	k1, v1 := "k1", String("v1")
	k2, v2 := "k2", String("v2")
	lru := NewLRU(int64(len(k1+k2)+v1.Len()+v2.Len()), nil)
	lru.Add(k1, v1)
	lru.Add(k2, v2)
	lru.Add("k3", String("v3"))
	if _, ok := lru.Get("k1"); ok || lru.Len() != 2 {
		t.Fatal("del oldest failed")
	}
}

func TestOnDeleted(t *testing.T) {
	k1, v1 := "k1", String("v1")
	k2, v2 := "k2", String("v2")
	var keys []string
	lru := NewLRU(int64(len(k1+k2)+v1.Len()+v2.Len()), func(key string, value Value) {
		keys = append(keys, key)
	})
	lru.Add(k1, v1)
	lru.Add(k2, v2)
	lru.Add("k3", String("v3"))
	if !reflect.DeepEqual([]string{"k1"}, keys) {
		t.Fatal("deleted func failed")
	}
}
