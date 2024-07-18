package strategy

import "container/list"

var _ Strategy = (*LRUK)(nil)

type LRUK struct {
	k         int
	capBytes  int64
	lenBytes  int64
	old       *list.List
	new       *list.List
	dataMap   map[string]*list.Element
	onDeleted func(key string, value Value)
}

type lruKEntry struct {
	key   string
	value Value
	count int
}

func NewLRUK(cap int64, on func(key string, value Value)) *LRUK {
	return &LRUK{
		k:        2,
		capBytes: cap,
		old:      list.New(),
		new:      list.New(),
		dataMap:  make(map[string]*list.Element),
	}
}

// Add implements Strategy.
func (l *LRUK) Add(key string, value Value) {
	ele, ok := l.dataMap[key]
	if ok {
		entry := ele.Value.(*lruKEntry)
		entry.value = value
		if entry.count == l.k-1 {
			l.old.Remove(ele)
			entry.count++
			l.new.PushFront(entry)
		} else if entry.count == l.k {
			l.new.MoveToFront(ele)
		}
		l.lenBytes += int64(value.Len()) - int64(entry.value.Len())
	} else {
		entry := &lruKEntry{
			key:   key,
			value: value,
			count: 1,
		}
		e := l.old.PushBack(entry)
		l.dataMap[key] = e
		l.lenBytes += int64(value.Len()) + int64(len(key))
	}

	for l.capBytes != 0 && l.lenBytes > l.capBytes {
		l.delOldest()
	}
}

// Get implements Strategy.
func (l *LRUK) Get(key string) (Value, bool) {
	ele, ok := l.dataMap[key]
	if ok {
		entry := ele.Value.(*lruKEntry)
		if entry.count == l.k-1 {
			entry.count++
			l.old.Remove(ele)
			l.new.PushFront(entry)
		} else if entry.count == l.k {
			l.new.MoveToFront(ele)
		} else {
			entry.count++
		}
		return entry.value, true
	} else {
		return nil, false
	}
}

func (l *LRUK) delOldest() {
	var ele *list.Element
	if l.new.Len() > 0 {
		ele = l.new.Back()
		l.new.Remove(ele)

	} else {
		ele = l.old.Front()
		l.old.Remove(ele)
	}
	entry := ele.Value.(*lruKEntry)
	delete(l.dataMap, entry.key)
	l.lenBytes -= int64(len(entry.key)) + int64(entry.value.Len())

	if l.onDeleted != nil {
		l.onDeleted(entry.key, entry.value)
	}
}

// Len implements Strategy.
func (l *LRUK) Len() int {
	return l.old.Len() + l.new.Len()
}
