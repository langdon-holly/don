package syntax

import "strings"

func subSyntaxString(out *strings.Builder, s Syntax, indent []byte, precedence SyntaxTag) {
	if s.Tag < precedence {
		out.WriteString("(")
		syntaxString(out, s, indent, false)
		out.WriteString(")")
	} else if syntaxString(out, s, indent, false); true {
	}
}
func syntaxString(out *strings.Builder, s Syntax, indent []byte, topLevel bool) {
	switch s.Tag {
	case ListSyntaxTag:
		subIndent := indent
		if !topLevel {
			subIndent = append(indent, byte('\t'))
		}
		out.WriteString("\n")
		for _, line := range s.Children {
			if line.Tag != EmptyLineSyntaxTag {
				out.Write(subIndent)
			}
			subSyntaxString(out, line, subIndent, ListSyntaxTag+1)
			out.WriteString("\n")
		}
		out.Write(indent)
	case EmptyLineSyntaxTag:
	case SpacedSyntaxTag:
		subSyntaxString(out, s.Children[0], indent, SpacedSyntaxTag+1)
		for i := 1; i < len(s.Children); i++ {
			out.WriteString(" ")
			subSyntaxString(out, s.Children[i], indent, SpacedSyntaxTag+1)
		}
	case MCallSyntaxTag:
		subSyntaxString(out, s.Children[0], indent, SandwichSyntaxTag+1)
		out.WriteString("!")
		subSyntaxString(out, s.Children[1], indent, MCallSyntaxTag)
	case SandwichSyntaxTag:
		subSyntaxString(out, s.Children[0], indent, SandwichSyntaxTag+1)
		out.WriteString("-")
		subSyntaxString(out, s.Children[1], indent, MCallSyntaxTag)
	case NameSyntaxTag:
		if s.LeftMarker {
			out.WriteString(":")
		}
		if s.Name != "" {
			for _, b := range []byte(s.Name) {
				if byteIsSpecial(b) {
					out.WriteString("\\")
				}
				out.WriteByte(b)
			}
		} else if out.WriteString("_"); true {
		}
		if s.RightMarker {
			out.WriteString(":")
		}
	}
	return
}
func (s Syntax) String() string {
	var b strings.Builder
	syntaxString(&b, s, nil, true)
	return b.String()
}

func EscapeFieldName(fieldName string) string {
	return Syntax{Tag: NameSyntaxTag, Name: fieldName}.String()
}
