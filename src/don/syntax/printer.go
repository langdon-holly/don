package syntax

import "strings"

func (s Syntax) subWriteString(out *strings.Builder, indent []byte, precedence SyntaxTag) {
	if s.Tag < precedence {
		out.WriteString("(")
		s.writeString(out, indent, false)
		out.WriteString(")")
	} else if s.writeString(out, indent, false); true {
	}
}
func (s Syntax) writeString(out *strings.Builder, indent []byte, topLevel bool) {
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
			line.subWriteString(out, subIndent, ListSyntaxTag+1)
			out.WriteString("\n")
		}
		out.Write(indent)
	case EmptyLineSyntaxTag:
	case ApplicationSyntaxTag:
		s.Children[0].subWriteString(out, indent, ApplicationSyntaxTag)
		out.WriteString(" ! ")
		s.Children[1].subWriteString(out, indent, ApplicationSyntaxTag+1)
	case CompositionSyntaxTag:
		s.Children[0].subWriteString(out, indent, CompositionSyntaxTag+1)
		for i := 1; i < len(s.Children); i++ {
			out.WriteString(" ")
			s.Children[i].subWriteString(out, indent, CompositionSyntaxTag+1)
		}
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
	case QuotationSyntaxTag:
		out.WriteString("{")
		s.Children[0].subWriteString(out, indent, 0)
		out.WriteString("}")
	}
	return
}
func (s Syntax) String() string {
	var b strings.Builder
	s.subWriteString(&b, nil, ListSyntaxTag+1)
	return b.String()
}

func (s Syntax) StringAtTop() string {
	var b strings.Builder
	s.writeString(&b, nil, true)
	return b.String()
}

func EscapeFieldName(fieldName string) string {
	return Syntax{Tag: NameSyntaxTag, Name: fieldName}.String()
}
