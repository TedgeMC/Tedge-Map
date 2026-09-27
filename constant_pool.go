package main

// ---- constant pool entry storage ----

type ConstantPoolEntry struct {
	tag                      int
	nameIndex                int
	classIndex               int
	nameAndTypeIndex         int
	stringIndex              int
	descriptorIndex          int
	referenceKind            int
	referenceIndex           int
	bootstrapMethodAttrIndex int
	intValue                 int32
	floatValue               float32
	longValue                int64
	doubleValue              float64
	utf8Value                string
}
