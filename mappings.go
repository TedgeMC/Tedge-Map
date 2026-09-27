package main

import "strings"

func transformLiteral(val string) string {
	i, ok := indexes[val]

	//NOTE: I know it looks ass but its the easiest way
	if ok {
		//fmt.Printf("%s :: ", val)
		//fmt.Printf("%d = %s\n", i, val)

		val = mappings[i:i+60]
		val = val[:strings.IndexRune(val, ' ')]
		val = strings.Replace(val, ".", "/", -1)
	}

	return val
}

func transformDescriptor(desc string) string {
	var out strings.Builder
	var value strings.Builder

	literal := false

	for _, ch := range desc {
		if literal {
			if ch == ';' {
				out.WriteRune('L')
				out.WriteString(transformLiteral(value.String()))
				out.WriteRune(';')

				value.Reset()
				literal = false
			} else {
				value.WriteRune(ch)
			}
		} else if ch == 'L' {
			literal = true
		} else {
			out.WriteRune(ch)
		}
	}

	return out.String()
}
