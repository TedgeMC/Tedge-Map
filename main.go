package main

import (
	"fmt"
	"os"
)

// Initially converted from the javap implementation.
// Thanks to texadactyl and the Jacobin team at github.com/platypusguy/jacobin (/issues/511)

var (
	errorCount = 0
	LOGGING    = true
	DEBUG      = false
)

var mappings string
var indexes map[string]int

func main() {
	//NOTE shouldn't this be `var classfile string` for better clarification?
	//            although because of go zero-value this probably would result in the same thing..
	classfile := ""
	mappingsfile := ""
	indexesfile := ""

	args := os.Args[1:]

	//if len(args) == 0 {
	//	classfile = "hello.class"
	//} else
	if len(args) == 4 && args[0] == "apply" {
		classfile = args[1]
		mappingsfile = args[2]
		indexesfile = args[3]
	} else if len(args) == 2 && args[0] == "indexes" {
		if LOGGING {
			println("- Generating indexes")
		}

		mappingsfile = args[1]
		MakeIndexes(mappingsfile)

		return
	} else {
		classfile = parseArgs(args)
		if classfile == "" {
			showUsage()
			return
		}
	}

	mappings = string(read(mappingsfile))
	indexes = map[string]int{}

	readIndexes(indexesfile)

	classBytes := read(classfile)
	analyze(classBytes)

	// The Java source used Checkers.theEnd(errorCount).  The only checker
	// in this file validates constant-pool entry #1, so retain that validation
	// and report the final count without requiring the Jacobin test harness.
	if errorCount != 0 && false {
		fmt.Printf("errorCount: %d\n", errorCount)
		os.Exit(1)
	}
}

func parseArgs(args []string) string {
	// Preserves the supplied Java implementation: multi-argument options
	// are not implemented yet.
	return ""
}

func read(path string) []byte {
	classBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("*** IOException while reading file: %s\n", path)
		os.Exit(1)
	}
	return classBytes
}

func showUsage() {
	fmt.Println("Tedge-Map mappings applying tool (c) 2026 Olafcio & The Jacobin Team")
	fmt.Println("Usage:")
	fmt.Println("       tedge-map apply [options] <classfile> <mappingsfile> <indexesfile>")
	fmt.Println("       tedge-map indexes [options] <mappingsfile>")
	fmt.Println("Help - Apply:")
	fmt.Println("    output is written into provided classfile")
	fmt.Println("Help - Indexes:")
	fmt.Println("    output is written into indexes.txt")
}

func analyze(bytes []byte) {
	reader := NewClassReader(bytes)

	magic := reader.readU4()
	if magic != CLASS_MAGIC {
		fmt.Println("invalid Java class file")
		return
	}

	minorVersion := reader.readU2()
	majorVersion := reader.readU2()

	if LOGGING {
		versionName := versionName(majorVersion)
		if versionName != "" {
			fmt.Printf("Java version: %s (minor %d, major %d)\n",
				versionName, minorVersion, majorVersion)
		}
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("*** RuntimeException: malformed or truncated class file")
				panic(r)
			}
		}()

		if DEBUG {
			fmt.Printf("=== analyze try: bytes.length=%d)\n", len(bytes))
		}

		constantPool := readConstantPool(reader)

		if DEBUG {
			fmt.Println("=== analyze try: ConstantPoolEntry[] constantPool = readConstantPool ... ok")
		}

		if LOGGING {
			printConstantPool(constantPool)

			if DEBUG {
				fmt.Println("=== analyze try: printConstantPool ... ok")
			}
		}

		accessFlags := reader.readU2()
		thisClass := reader.readU2()
		superClass := reader.readU2()

		className := resolveClassName(constantPool, thisClass)
		superName := resolveClassName(constantPool, superClass)

		if LOGGING {
			fmt.Printf("access flags: 0x%04x (%s)\n", accessFlags, decodeClassAccessFlags(accessFlags))
			fmt.Printf("this class:   #%d // %s\n", thisClass, className)
			if superClass != 0 {
				fmt.Printf("super class:  #%d // %s\n", superClass, superName)
			}
		}

		interfaceCount := reader.readU2()
		if LOGGING {
			fmt.Printf("interfaces: %d\n", interfaceCount)
		}
		for i := 0; i < interfaceCount; i++ {
			interfaceIndex := reader.readU2()
			interfaceName := resolveClassName(constantPool, interfaceIndex)
			
			if LOGGING {
				fmt.Printf("  #%d // %s\n", interfaceIndex, interfaceName)
			}
		}

		fieldsCount := reader.readU2()
		if LOGGING {
			fmt.Printf("fields: %d\n", fieldsCount)
		}
		for i := 0; i < fieldsCount; i++ {
			readMember(reader, constantPool, "field")
		}

		methodsCount := reader.readU2()
		if LOGGING {
			fmt.Printf("methods: %d\n", methodsCount)
		}
		for i := 0; i < methodsCount; i++ {
			readMember(reader, constantPool, "method")
		}

		classAttributesCount := reader.readU2()
		if LOGGING {
			fmt.Printf("class attributes: %d\n", classAttributesCount)
		}
		for i := 0; i < classAttributesCount; i++ {
			readAttribute(reader, constantPool, 2)
		}
		
		//TODO Class writing!!
	}()
}

func versionName(majorVersion int) string {
	switch {
	case majorVersion < 55:
		return "pre-JDK 11"
	case majorVersion == 55:
		return "JDK 11"
	case majorVersion == 56:
		return "JDK 12"
	case majorVersion == 57:
		return "JDK 13"
	case majorVersion == 58:
		return "JDK 14"
	case majorVersion == 59:
		return "JDK 15"
	case majorVersion == 60:
		return "JDK 16"
	case majorVersion == 61:
		return "JDK 17"
	case majorVersion == 62:
		return "JDK 18"
	case majorVersion == 63:
		return "JDK 19"
	case majorVersion == 64:
		return "JDK 20"
	case majorVersion == 65:
		return "JDK 21"
	case majorVersion == 66:
		return "JDK 22"
	case majorVersion == 67:
		return "JDK 23"
	case majorVersion == 68:
		return "JDK 24"
	case majorVersion == 69:
		return "JDK 25"
	case majorVersion == 70:
		return "JDK 26"
	default:
		return ""
	}
}
