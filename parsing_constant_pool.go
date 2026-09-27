package main

import "strings"
import "math"
import "fmt"

// ---- constant pool parsing ----

func readConstantPool(reader *ClassReader) []*ConstantPoolEntry {
	constantPoolCount := reader.readU2()
	pool := make([]*ConstantPoolEntry, constantPoolCount)

	for i := 1; i < constantPoolCount; i++ {
		tag := reader.readU1()
		entry := &ConstantPoolEntry{tag: tag}

		switch tag {
		case CONSTANT_UTF8:
			entry.utf8Value = reader.readUtf8()
		case CONSTANT_INTEGER:
			entry.intValue = int32(reader.readU4())
		case CONSTANT_FLOAT:
			entry.floatValue = math.Float32frombits(reader.readU4())
		case CONSTANT_LONG:
			high := uint64(reader.readU4())
			low := uint64(reader.readU4())
			entry.longValue = int64((high << 32) | low)
			pool[i] = entry
			i++
			continue
		case CONSTANT_DOUBLE:
			high := uint64(reader.readU4())
			low := uint64(reader.readU4())
			entry.doubleValue = math.Float64frombits((high << 32) | low)
			pool[i] = entry
			i++
			continue
		case CONSTANT_CLASS, CONSTANT_METHOD_TYPE, CONSTANT_MODULE, CONSTANT_PACKAGE:
			entry.nameIndex = reader.readU2()
		case CONSTANT_STRING:
			entry.stringIndex = reader.readU2()
		case CONSTANT_FIELDREF, CONSTANT_METHODREF, CONSTANT_INTERFACE_METHODREF:
			entry.classIndex = reader.readU2()
			entry.nameAndTypeIndex = reader.readU2()
		case CONSTANT_NAME_AND_TYPE:
			entry.nameIndex = reader.readU2()
			entry.descriptorIndex = reader.readU2()
		case CONSTANT_METHOD_HANDLE:
			entry.referenceKind = reader.readU1()
			entry.referenceIndex = reader.readU2()
		case CONSTANT_DYNAMIC, CONSTANT_INVOKE_DYNAMIC:
			entry.bootstrapMethodAttrIndex = reader.readU2()
			entry.nameAndTypeIndex = reader.readU2()
		default:
			panic(fmt.Sprintf("unknown constant pool tag %d at index %d", tag, i))
		}

		pool[i] = entry
	}

	return pool
}

func printConstantPool(pool []*ConstantPoolEntry) {
	fmt.Println("Constant pool:")
	for i := 1; i < len(pool); i++ {
		entry := pool[i]
		if entry == nil {
			continue
		}

		fpe := formatConstantPoolEntry(pool, entry)
		fmt.Printf("  #%-4d= %s\n", i, fpe)
		if i == 1 {
			expected := "Methodref          #2.#3 // java/lang/Object.<init>:()V"
			if fpe != expected {
				errorCount++
				fmt.Printf("*** checker failed: CP entry #1: expected %q, got %q\n", expected, fpe)
			}
		}
	}
}

func formatConstantPoolEntry(pool []*ConstantPoolEntry, entry *ConstantPoolEntry) string {
	if DEBUG {
		fmt.Printf("=== formatConstantPoolEntry begin: entry.tag=%d, entry.nameIndex=%d, entry.stringIndex=%d, entry.descriptorIndex=%d, entry.nameAndTypeIndex=%d, pool.length=%d\n",
			entry.tag, entry.nameIndex, entry.stringIndex, entry.descriptorIndex, entry.nameAndTypeIndex, len(pool))
	}

	switch entry.tag {
	case CONSTANT_UTF8:
		return fmt.Sprintf("Utf8               %s", entry.utf8Value)
	case CONSTANT_INTEGER:
		return fmt.Sprintf("Integer            %d", entry.intValue)
	case CONSTANT_FLOAT:
		return fmt.Sprintf("Float              %f", entry.floatValue)
	case CONSTANT_LONG:
		return fmt.Sprintf("Long               %d", entry.longValue)
	case CONSTANT_DOUBLE:
		return fmt.Sprintf("Double             %f", entry.doubleValue)
	case CONSTANT_CLASS:
		return fmt.Sprintf("Class              #%d // %s", entry.nameIndex, utf8At(pool, entry.nameIndex))
	case CONSTANT_STRING:
		return fmt.Sprintf("String             #%d // %s", entry.stringIndex, utf8At(pool, entry.stringIndex))
	case CONSTANT_FIELDREF:
		return fmt.Sprintf("Fieldref           #%d.#%d // %s", entry.classIndex, entry.nameAndTypeIndex, describeRef(pool, entry))
	case CONSTANT_METHODREF:
		return fmt.Sprintf("Methodref          #%d.#%d // %s", entry.classIndex, entry.nameAndTypeIndex, describeRef(pool, entry))
	case CONSTANT_INTERFACE_METHODREF:
		return fmt.Sprintf("InterfaceMethodref #%d.#%d // %s", entry.classIndex, entry.nameAndTypeIndex, describeRef(pool, entry))
	case CONSTANT_NAME_AND_TYPE:
		return fmt.Sprintf("NameAndType        #%d:#%d // %s:%s",
			entry.nameIndex, entry.descriptorIndex,
			utf8At(pool, entry.nameIndex), utf8At(pool, entry.descriptorIndex))
	case CONSTANT_METHOD_HANDLE:
		return fmt.Sprintf("MethodHandle       kind=%d #%d", entry.referenceKind, entry.referenceIndex)
	case CONSTANT_METHOD_TYPE:
		return fmt.Sprintf("MethodType         #%d // %s", entry.nameIndex, utf8At(pool, entry.nameIndex))
	case CONSTANT_DYNAMIC:
		return fmt.Sprintf("Dynamic            #%d:#%d", entry.bootstrapMethodAttrIndex, entry.nameAndTypeIndex)
	case CONSTANT_INVOKE_DYNAMIC:
		return fmt.Sprintf("InvokeDynamic      #%d:#%d", entry.bootstrapMethodAttrIndex, entry.nameAndTypeIndex)
	case CONSTANT_MODULE:
		return fmt.Sprintf("Module             #%d // %s", entry.nameIndex, utf8At(pool, entry.nameIndex))
	case CONSTANT_PACKAGE:
		return fmt.Sprintf("Package            #%d // %s", entry.nameIndex, utf8At(pool, entry.nameIndex))
	default:
		return fmt.Sprintf("Unknown tag %d", entry.tag)
	}
}

func utf8At(pool []*ConstantPoolEntry, index int) string {
	if DEBUG {
		fmt.Printf("=== utf8At begin: index=%d, pool.length=%d\n", index, len(pool))
	}
	if index <= 0 || index >= len(pool) || pool[index] == nil {
		panic(fmt.Sprintf("*** utf8At index: %d, pool length: %d", index, len(pool)))
	}
	if DEBUG {
		fmt.Printf("=== utf8At end: pool[index=%d] %#v\n", index, pool[index])
	}
	return pool[index].utf8Value
}

func describeRef(pool []*ConstantPoolEntry, entry *ConstantPoolEntry) string {
	className := resolveClassName(pool, entry.classIndex)
	if DEBUG {
		fmt.Printf("=== describeRef begin: className=%s\n", className)
	}
	if entry.nameAndTypeIndex <= 0 || entry.nameAndTypeIndex >= len(pool) || pool[entry.nameAndTypeIndex] == nil {
		return fmt.Sprintf("%s.? ", className)
	}
	nameAndType := pool[entry.nameAndTypeIndex]
	name := utf8At(pool, nameAndType.nameIndex)
	descriptor := utf8At(pool, nameAndType.descriptorIndex)
	return fmt.Sprintf("%s.%s:%s", className, name, descriptor)
}

func resolveClassName(pool []*ConstantPoolEntry, classIndex int) string {
	if DEBUG {
		fmt.Printf("=== resolveClassName begin: classIndex=%d\n", classIndex)
	}
	if classIndex <= 0 || classIndex >= len(pool) || pool[classIndex] == nil {
		return "?"
	}
	if DEBUG {
		fmt.Printf("=== resolveClassName end: pool[classIndex].nameIndex=%d\n", pool[classIndex].nameIndex)
	}
	return utf8At(pool, pool[classIndex].nameIndex)
}

func decodeClassAccessFlags(flags int) string {
	var names []string
	if flags&0x0001 != 0 {
		names = append(names, "public")
	}
	if flags&0x0010 != 0 {
		names = append(names, "final")
	}
	if flags&0x0020 != 0 {
		names = append(names, "super")
	}
	if flags&0x0200 != 0 {
		names = append(names, "interface")
	}
	if flags&0x0400 != 0 {
		names = append(names, "abstract")
	}
	if flags&0x1000 != 0 {
		names = append(names, "synthetic")
	}
	if flags&0x2000 != 0 {
		names = append(names, "annotation")
	}
	if flags&0x4000 != 0 {
		names = append(names, "enum")
	}
	if flags&0x8000 != 0 {
		names = append(names, "module")
	}
	// The Java source intended to concatenate the names but used String.concat
	// without assigning the result.  Use the intended output here.
	return strings.Join(names, " ")
}
