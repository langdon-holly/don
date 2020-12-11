package syntax

import "strings"

func (s Syntax) String() string {
	switch s.Tag {
	case NameSyntaxTag:
		var builder strings.Builder
		if s.LeftMarker {
			builder.WriteString(":")
		}
		if s.Name != "" {
			for _, b := range []byte(s.Name) {
				if byteIsSpecial(b) {
					builder.WriteString("\\")
				}
				builder.WriteByte(b)
			}
		} else if builder.WriteString("_"); true {
		}
		if s.RightMarker {
			builder.WriteString(":")
		}
		return builder.String()
	default:
		panic("Unimplemented")
	}
	panic("Unreachable")
}

func EscapeFieldName(fieldName string) string {
	return Syntax{Tag: NameSyntaxTag, Name: fieldName}.String()
}
