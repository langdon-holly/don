package syntax

import "strings"

const MAX_TOKENS_PER_LINE = 16 /* Suggestion */

type layoutInfo struct {
	Circumfix bool
	Tokens    []int /* 1 or 2 elements */
}

func (l layoutInfo) Top() int { return l.Tokens[0] }
func (l layoutInfo) Bot() int { return l.Tokens[len(l.Tokens)-1] }

func (l0 layoutInfo) Compose(tokens int, l1 layoutInfo) layoutInfo {
	top := l0.Top()
	bot := l1.Bot()
	if len(l0.Tokens) == 1 {
		top = top + tokens + l1.Top()
	} else if len(l1.Tokens) == 1 {
		bot = bot + tokens + l0.Bot()
	}
	if len(l0.Tokens) == 1 && len(l1.Tokens) == 1 {
		return layoutInfo{Circumfix: l0.Circumfix || l1.Circumfix, Tokens: []int{top}}
	} else {
		return layoutInfo{
			Circumfix: l0.Circumfix || l1.Circumfix,
			Tokens:    []int{top, bot},
		}
	}
}

func (s Syntax) subLayout(precedence SyntaxTag) (l layoutInfo, writeString func(out *strings.Builder, indent []byte)) {
	if subL, subWriteString := s.layout(); s.Tag >= precedence {
		l, writeString = subL, subWriteString
	} else if len(subL.Tokens) == 2 ||
		subL.Circumfix ||
		subL.Top()+1 > MAX_TOKENS_PER_LINE {
		l = layoutInfo{Circumfix: true, Tokens: []int{1, 1}}
		writeString = func(out *strings.Builder, indent []byte) {
			subIndent := append(indent, "\t"...)
			out.WriteString("(")
			out.WriteString("\n")
			out.Write(subIndent)
			subWriteString(out, subIndent)
			out.WriteString("\n")
			out.Write(indent)
			out.WriteString(")")
		}
	} else {
		l = layoutInfo{Circumfix: true, Tokens: []int{subL.Top() + 2}}
		writeString = func(out *strings.Builder, indent []byte) {
			out.WriteString("(")
			subWriteString(out, indent)
			out.WriteString(")")
		}
	}
	return
}
func (s Syntax) layout() (l layoutInfo, writeString func(out *strings.Builder, indent []byte)) {
	switch s.Tag {
	case ListSyntaxTag:
		if len(s.Children) == 0 {
			l.Tokens = []int{1}
			writeString = func(out *strings.Builder, _ []byte) { out.WriteString("*") }
		} else if len(s.Children) == 1 {
			var factorWriteString func(*strings.Builder, []byte)
			l, factorWriteString = s.Children[0].subLayout(ListSyntaxTag + 1)
			l.Tokens[0]++
			writeString = func(out *strings.Builder, indent []byte) {
				out.WriteString("* ")
				factorWriteString(out, indent)
			}
		} else if factorEmptyLines := make([]bool, len(s.Children)); true {
			factorLs := make([]layoutInfo, len(s.Children))
			factorWriteStrings := make([]func(*strings.Builder, []byte), len(s.Children))
			multiline := false
			tokens := 0
			for i, factor := range s.Children {
				factorEmptyLines[i] = factor.Tag == EmptyLineSyntaxTag
				factorLs[i], factorWriteStrings[i] = factor.subLayout(ListSyntaxTag + 1)
				multiline = multiline || len(factorLs[i].Tokens) == 2
				l.Circumfix = l.Circumfix || factorLs[i].Circumfix
				tokens += factorLs[i].Top() + 1
			}
			tokens--
			if multiline || tokens > MAX_TOKENS_PER_LINE {
				l.Tokens = []int{
					factorLs[0].Top() + 1,
					factorLs[len(factorLs)-1].Bot() + 1,
				}
				writeString = func(out *strings.Builder, indent []byte) {
					out.WriteString("* ")
					factorWriteStrings[0](out, indent)
					for i := 1; i < len(factorWriteStrings); i++ {
						if out.WriteString("\n"); !factorEmptyLines[i] {
							out.Write(indent)
							out.WriteString("* ")
							factorWriteStrings[i](out, indent)
						}
					}
				}
			} else if l.Tokens = []int{tokens}; true {
				writeString = func(out *strings.Builder, indent []byte) {
					factorWriteStrings[0](out, indent)
					for i := 1; i < len(factorWriteStrings); i++ {
						out.WriteString(" * ")
						factorWriteStrings[i](out, indent)
					}
				}
			}
		}
	case EmptyLineSyntaxTag:
		l.Tokens = []int{0, 0}
		writeString = func(*strings.Builder, []byte) { panic("Unreachable") }
	case ApplicationSyntaxTag:
		comL, comWriteString := s.Children[0].subLayout(ApplicationSyntaxTag)
		argL, argWriteString := s.Children[1].subLayout(ApplicationSyntaxTag + 1)
		l = comL.Compose(1, argL)
		writeString = func(out *strings.Builder, indent []byte) {
			comWriteString(out, indent)
			out.WriteString(" ! ")
			argWriteString(out, indent)
		}
	case CompositionSyntaxTag:
		l.Tokens = []int{0}
		factorWriteStrings := make([]func(*strings.Builder, []byte), len(s.Children))
		for i, factor := range s.Children {
			var factorL layoutInfo
			factorL, factorWriteStrings[i] = factor.subLayout(CompositionSyntaxTag + 1)
			l = l.Compose(0, factorL)
		}
		writeString = func(out *strings.Builder, indent []byte) {
			factorWriteStrings[0](out, indent)
			for i := 1; i < len(factorWriteStrings); i++ {
				out.WriteString(" ")
				factorWriteStrings[i](out, indent)
			}
		}
	case NameSyntaxTag:
		l.Tokens = []int{1}
		writeString = func(out *strings.Builder, _ []byte) {
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
	case ISyntaxTag:
		l.Tokens = []int{2}
		writeString = func(out *strings.Builder, _ []byte) { out.WriteString("()") }
	case QuotationSyntaxTag:
		subL, subWriteString := s.Children[0].subLayout(0)
		if len(subL.Tokens) == 2 ||
			subL.Circumfix ||
			subL.Top()+1 > MAX_TOKENS_PER_LINE {
			l = layoutInfo{Circumfix: true, Tokens: []int{1, 1}}
			writeString = func(out *strings.Builder, indent []byte) {
				subIndent := append(indent, "\t"...)
				out.WriteString("{")
				out.WriteString("\n")
				out.Write(subIndent)
				subWriteString(out, subIndent)
				out.WriteString("\n")
				out.Write(indent)
				out.WriteString("}")
			}
		} else {
			l = layoutInfo{Circumfix: true, Tokens: []int{subL.Top() + 2}}
			writeString = func(out *strings.Builder, indent []byte) {
				out.WriteString("{")
				subWriteString(out, indent)
				out.WriteString("}")
			}
		}
	default:
		panic("Unreachable")
	}
	return
}

func (s Syntax) String() string {
	_, writeString := s.subLayout(ListSyntaxTag + 1)
	var b strings.Builder
	writeString(&b, nil)
	return b.String()
}

func (s Syntax) TopString() string {
	_, writeString := s.Children[1].Children[0].layout()
	var b strings.Builder
	writeString(&b, nil)
	return b.String()
}

func EscapeFieldName(fieldName string) string {
	return Syntax{Tag: NameSyntaxTag, Name: fieldName}.String()
}
