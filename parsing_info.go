package main

import "fmt"
import "strings"
import "encoding/binary"

// ---- field_info / method_info / attribute_info traversal ----

func readMember(reader *ClassReader, pool []*ConstantPoolEntry, typ string) {
	accessFlags := reader.readU2()
	_ = accessFlags
	nameIndex := reader.readU2()
	descriptorIndex := reader.readU2()
	attributesCount := reader.readU2()

	if LOGGING {
		fmt.Printf("  %s %s %s\n", typ, utf8At(pool, nameIndex), utf8At(pool, descriptorIndex))
	}

	for i := 0; i < attributesCount; i++ {
		readAttribute(reader, pool, 4)
	}
}

func readAttribute(reader *ClassReader, pool []*ConstantPoolEntry, indent int) {
	nameIndex := reader.readU2()
	length := reader.readU4()
	name := utf8At(pool, nameIndex)
	indentStr := strings.Repeat(" ", indent)

	if LOGGING {
		fmt.Printf("%sAttribute: %s (length: %d)\n", indentStr, name, length)
	}

	if name == "Code" {
		maxStack := reader.readU2()
		maxLocals := reader.readU2()
		codeLength := reader.readU4()
		code := reader.readBytes(codeLength)

		if LOGGING {
			fmt.Printf("%s  stack=%d, locals=%d, code_length=%d\n", indentStr, maxStack, maxLocals, codeLength)
			decodeInstructions(code, pool, indent+4)
		}

		exceptionTableLength := reader.readU2()
		if exceptionTableLength > 0 && LOGGING {
			fmt.Printf("%s  Exception table:\n", indentStr)
			fmt.Printf("%s     from    to  target type\n", indentStr)
		}
		for i := 0; i < exceptionTableLength; i++ {
			start := reader.readU2()
			end := reader.readU2()
			handler := reader.readU2()
			catchType := reader.readU2()
			if LOGGING {
				typeStr := "any"
				if catchType != 0 {
					typeStr = resolveClassName(pool, catchType)
				}
				fmt.Printf("%s     %4d  %4d  %4d   %s\n", indentStr, start, end, handler, typeStr)
			}
		}

		attributesCount := reader.readU2()
		for i := 0; i < attributesCount; i++ {
			readAttribute(reader, pool, indent+2)
		}
	} else if name == "LineNumberTable" && LOGGING {
		lineTableLength := reader.readU2()
		for i := 0; i < lineTableLength; i++ {
			startPC := reader.readU2()
			lineNumber := reader.readU2()
			fmt.Printf("%s  line %d: %d\n", indentStr, lineNumber, startPC)
		}
	} else {
		reader.skip(int(length))
	}
}

func decodeInstructions(code []byte, pool []*ConstantPoolEntry, indent int) {
	indentStr := strings.Repeat(" ", indent)
	pc := 0

	for pc < len(code) {
		opcode := int(code[pc])
		name := opcodeName(opcode)
		length := opcodeLength(opcode, code, pc)

		var sb strings.Builder
		fmt.Fprintf(&sb, "%s%4d: %-15s", indentStr, pc, name)

		comment := ""
		if length > 1 && pc+length <= len(code) {
			switch opcode {
			case 0x12:
				index := int(code[pc+1])
				fmt.Fprintf(&sb, "#%-18d", index)
				comment = formatConstantPoolComment(pool, index)
			case 0x13, 0x14:
				index := int(binary.BigEndian.Uint16(code[pc+1 : pc+3]))
				fmt.Fprintf(&sb, "#%-18d", index)
				comment = formatConstantPoolComment(pool, index)
			case 0x10:
				val := int(int8(code[pc+1]))
				fmt.Fprintf(&sb, "%-19d", val)
			case 0x11:
				val := int(int16(binary.BigEndian.Uint16(code[pc+1 : pc+3])))
				fmt.Fprintf(&sb, "%-19d", val)
			case 0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xb8,
				0xbb, 0xbd, 0xc0, 0xc1:
				index := int(binary.BigEndian.Uint16(code[pc+1 : pc+3]))
				fmt.Fprintf(&sb, "#%-18d", index)
				comment = formatConstantPoolComment(pool, index)
			case 0xb9, 0xba:
				index := int(binary.BigEndian.Uint16(code[pc+1 : pc+3]))
				fmt.Fprintf(&sb, "#%-18d", index)
				comment = formatConstantPoolComment(pool, index)
			case 0x99, 0x9a, 0x9b, 0x9c, 0x9d, 0x9e,
				0x9f, 0xa0, 0xa1, 0xa2, 0xa3, 0xa4,
				0xa5, 0xa6, 0xa7, 0xa8, 0xc6, 0xc7:
				offset := int(int16(binary.BigEndian.Uint16(code[pc+1 : pc+3])))
				fmt.Fprintf(&sb, "%-19d", pc+offset)
			case 0xc8, 0xc9:
				offset := int(int32(binary.BigEndian.Uint32(code[pc+1 : pc+5])))
				fmt.Fprintf(&sb, "%-19d", pc+offset)
			case 0x84:
				index := int(code[pc+1])
				constVal := int(int8(code[pc+2]))
				fmt.Fprintf(&sb, "%d, %-16d", index, constVal)
			default:
				sb.WriteString(" // ")
				for i := 1; i < length; i++ {
					fmt.Fprintf(&sb, "%02x ", code[pc+i])
				}
			}
		}

		if comment != "" {
			sb.WriteString(" // ")
			sb.WriteString(comment)
		}
		fmt.Println(strings.TrimRight(sb.String(), " \t"))
		pc += length
	}
}

func formatConstantPoolComment(pool []*ConstantPoolEntry, index int) string {
	if index <= 0 || index >= len(pool) || pool[index] == nil {
		return ""
	}

	entry := pool[index]
	switch entry.tag {
	case CONSTANT_CLASS:
		return "Class " + resolveClassName(pool, index)
	case CONSTANT_STRING:
		return "String " + utf8At(pool, entry.stringIndex)
	case CONSTANT_FIELDREF:
		return "Field " + describeRef(pool, entry)
	case CONSTANT_METHODREF:
		return "Method " + describeRef(pool, entry)
	case CONSTANT_INTERFACE_METHODREF:
		return "InterfaceMethod " + describeRef(pool, entry)
	case CONSTANT_NAME_AND_TYPE:
		return "NameAndType " + utf8At(pool, entry.nameIndex) + ":" + utf8At(pool, entry.descriptorIndex)
	case CONSTANT_INTEGER:
		return fmt.Sprintf("int %d", entry.intValue)
	case CONSTANT_FLOAT:
		return fmt.Sprintf("float %gf", entry.floatValue)
	case CONSTANT_LONG:
		return fmt.Sprintf("long %dl", entry.longValue)
	case CONSTANT_DOUBLE:
		return fmt.Sprintf("double %gd", entry.doubleValue)
	case CONSTANT_UTF8:
		return "Utf8 " + entry.utf8Value
	case CONSTANT_INVOKE_DYNAMIC:
		return fmt.Sprintf("InvokeDynamic #%d:#%d", entry.bootstrapMethodAttrIndex, entry.nameAndTypeIndex)
	default:
		return ""
	}
}
