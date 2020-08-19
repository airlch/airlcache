package airlcache

type ByteView struct {
	b []byte
}

func (by *ByteView) Len() int {
	return len(by.b)
}

func (by *ByteView) String() string {
	return string(by.b)
}

func (by *ByteView) ByteSlice() []byte {
	return cloneBytes(by.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
