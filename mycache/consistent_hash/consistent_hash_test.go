package consistent_hash

import (
	"strconv"
	"testing"
)

func TestMap(t *testing.T) {
	m := NewHash(3, func(data []byte) uint32 {
		v, _ := strconv.Atoi(string(data))
		return uint32(v)
	})

	m.AddNode("1", "4", "7")
	cases := map[string]string{
		"0":  "1",
		"2":  "4",
		"15": "7",
		"31": "1",
	}

	for k, v := range cases {
		if m.GetNode(k) != v {
			t.Fatalf("ask for %s, expected %s, get %s", k, v, m.GetNode(k))
		}
	}

	m.AddNode("5")
	cases["15"] = "5"
	for k, v := range cases {
		if m.GetNode(k) != v {
			t.Fatalf("ask for %s, expected %s, get %s", k, v, m.GetNode(k))
		}
	}
}
