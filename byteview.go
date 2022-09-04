package gdcache

// ByteView 一个只读数据结构，用来封装缓存值
type ByteView struct {
	b []byte
}

// Len method 获取缓存值字节长度
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice method 返回一个拷贝的字节切片
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)

	return c
}
