package main

// ---- low-level byte cursor over the class file ----

type ClassReader struct {
	data     []byte
	position int
}

func NewClassReader(data []byte) *ClassReader {
	return &ClassReader{data: data}
}

func (r *ClassReader) readU1() int {
	v := r.data[r.position]
	r.position++
	return int(v)
}

func (r *ClassReader) readU2() int {
	high := r.readU1()
	low := r.readU1()
	return (high << 8) | low
}

func (r *ClassReader) readU4() uint32 {
	b1 := r.readU1()
	b2 := r.readU1()
	b3 := r.readU1()
	b4 := r.readU1()
	return uint32(b1)<<24 | uint32(b2)<<16 | uint32(b3)<<8 | uint32(b4)
}

func (r *ClassReader) readBytes(n uint32) []byte {
	start := r.position
	end := start + int(n)
	bytes := make([]byte, int(n))
	copy(bytes, r.data[start:end])
	r.position = end
	return bytes
}

func (r *ClassReader) readUtf8() string {
	length := r.readU2()
	bytes := make([]byte, length)
	copy(bytes, r.data[r.position:r.position+length])
	r.position += length
	return string(bytes)
}

func (r *ClassReader) skip(n int) {
	r.position += n
}
