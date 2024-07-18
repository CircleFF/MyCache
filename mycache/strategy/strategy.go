package strategy

type Strategy interface {
	Get(key string) (Value, bool)
	Add(key string, value Value)
	Len() int
}

type Value interface {
	Len() int
}

func New(name string, cap int64, on func(key string, value Value), k ...int) Strategy {
	switch name {
	case "lru":
		return NewLRU(cap, on)
	case "lfu":
		return NewLFU(cap, on)
	case "kruk":
		return NewLRUK(cap, on)
	}
	return nil
}
