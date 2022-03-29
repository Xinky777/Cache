package Cache

//ByteView 拥有一个不可变的字节视图
type ByteView struct {
	b []byte
}

//Len 返回view的长度
func (v ByteView) Len() int {
	return len(v.b)
}

//ByteSlice 以字节切片的形式返回数据的副本
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

//String 以字符串的形式返回数据，如果有必要的话进行副本
func (v ByteView) String() string {
	return string(v.b)
}
