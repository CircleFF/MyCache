package strategy

import "testing"

var lrukTestData = map[string]String{
	"k1": "v1",
	"k2": "v2",
	"k3": "v3",
}

func TestLRUK(t *testing.T) {
	lruk := NewLRUK(12, nil)
	for k, v := range lrukTestData {
		lruk.Add(k, v)
	}
	if lruk.new.Len() != 0 || lruk.old.Len() == 0 {
		t.Fatalf("length of new should be zero or old shouln't be zero, new:%d, old:%d", lruk.new.Len(), lruk.old.Len())
	}

	if val, ok := lruk.Get("k1"); !ok || val.(String) != "v1" || lruk.new.Len() != 1 {
		t.Fatal("get failed, k1 should exist or new should be 1")
	}

	lruk.Add("k4", String("v4"))
	if _, ok := lruk.Get("k1"); ok || lruk.old.Len() != 3 || lruk.new.Len() != 0 {
		t.Fatal("delete k1 failed")
	}
}
