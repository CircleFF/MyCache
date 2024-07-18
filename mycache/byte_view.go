package mycache

import "MyCache/mycache/strategy"

var _ strategy.Value = (*ByteView)(nil)

// ByteView 包装字节切片，实现 lru.Value 接口。
type ByteView struct {
	b []byte
}

func (B ByteView) Len() int {
	return len(B.b)
}

// String 返回底层字节切片的字符串表示
func (B ByteView) String() string {
	return string(B.b)
}

// ByteSlice 返回底层字节切片副本，防止原始数据被修改。
func (B ByteView) ByteSlice() []byte {
	return cloneBytes(B.b)
}

// cloneBytes 返回字节切片的副本
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
