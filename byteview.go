package comical_kv

// ByteView holds an immutable view of bytes
type ByteView struct {
	// b is a read-only slice
	b []byte
}

// Len returns the view's length, implementing the Value interface
func (bv ByteView) Len() int {
	return len(bv.b)
}

// ByteSlice returns a copy of the data as a byte slice
func (bv ByteView) ByteSlice() []byte {
	return cloneBytes(bv.b)
}

// String returns the data as a string
func (bv ByteView) String() string {
	return string(bv.b)
}

// cloneBytes returns a copy of the byte slice
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
