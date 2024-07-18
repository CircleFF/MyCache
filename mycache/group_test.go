package mycache

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	if ret, _ := f.Get("key"); !reflect.DeepEqual(ret, []byte("key")) {
		t.Fatal("failed")
	}
}

var db = map[string]string{
	"k1": "v1",
	"k2": "v2",
	"k3": "v3",
}

func TestGroup(t *testing.T) {
	loadCounts := make(map[string]int)
	group := NewGroup("mygroup", "lru", 1<<10, GetterFunc(func(key string) ([]byte, error) {
		if val, ok := db[key]; !ok {
			return nil, fmt.Errorf("local not have %s", key)
		} else {
			loadCounts[key]++
			return []byte(val), nil
		}
	}))

	for k, v := range db {
		if bv, err := group.Get(k); err != nil || bv.String() != v {
			t.Fatalf("mycache miss %s", k)
		}
		if _, err := group.Get(k); err != nil || loadCounts[k] != 1 {
			t.Fatalf("mycache %s failed", k)
		}
	}

	if bv, err := group.Get("empty"); err == nil || bv.Len() != 0 {
		t.Fatalf("key=%s should't exist", "empty")
	}
}
