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

func (List) precedence() int        { return ListPrecedence }
func (EmptyLine) precedence() int   { return EmptyLinePrecedence }
func (Application) precedence() int { return ApplicationPrecedence }
func (Composition) precedence() int { return CompositionPrecedence }
func (Named) precedence() int       { return NamedPrecedence }
func (ISyntax) precedence() int     { return ISyntaxPrecedence }
func (Quote) precedence() int       { return QuotePrecedence }

func subLayout(s Syntax, precedence int) (l layoutInfo, writeString func(out *strings.Builder, indent []byte)) {
	if subL, subWriteString := s.layout(); s.precedence() >= precedence {
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
func (list List) layout() (l layoutInfo, writeString func(out *strings.Builder, indent []byte)) {
	if len(list.Factors) == 0 {
		l.Tokens = []int{1}
		writeString = func(out *strings.Builder, _ []byte) { out.WriteString("*") }
	} else if len(list.Factors) == 1 {
		var factorWriteString func(*strings.Builder, []byte)
		l, factorWriteString = subLayout(list.Factors[0], ListPrecedence+1)
		l.Tokens[0]++
		writeString = func(out *strings.Builder, indent []byte) {
			out.WriteString("* ")
			factorWriteString(out, indent)
		}
	} else if factorEmptyLines := make([]bool, len(list.Factors)); true {
		factorLs := make([]layoutInfo, len(list.Factors))
		factorWriteStrings := make([]func(*strings.Builder, []byte), len(list.Factors))
		multiline := false
		tokens := 0
		for i, factor := range list.Factors {
			_, factorEmptyLines[i] = factor.(EmptyLine)
			factorLs[i], factorWriteStrings[i] = subLayout(factor, ListPrecedence+1)
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
	return
}
func (EmptyLine) layout() (l layoutInfo, writeString func(out *strings.Builder, indent []byte)) {
	l.Tokens = []int{0, 0}
	writeString = func(*strings.Builder, []byte) { panic("Unreachable") }
	return
}
func (a Application) layout() (l layoutInfo, writeString func(out *strings.Builder, indent []byte)) {
	comL, comWriteString := subLayout(a.Com, ApplicationPrecedence)
	argL, argWriteString := subLayout(a.Arg, ApplicationPrecedence+1)
	l = comL.Compose(1, argL)
	writeString = func(out *strings.Builder, indent []byte) {
		comWriteString(out, indent)
		out.WriteString(" ! ")
		argWriteString(out, indent)
	}
	return
}
func (c Composition) layout() (l layoutInfo, writeString func(out *strings.Builder, indent []byte)) {
	l.Tokens = []int{0}
	factorWriteStrings := make([]func(*strings.Builder, []byte), len(c.Factors))
	for i, factor := range c.Factors {
		var factorL layoutInfo
		factorL, factorWriteStrings[i] = subLayout(factor, CompositionPrecedence+1)
		l = l.Compose(0, factorL)
	}
	writeString = func(out *strings.Builder, indent []byte) {
		factorWriteStrings[0](out, indent)
		for i := 1; i < len(factorWriteStrings); i++ {
			out.WriteString(" ")
			factorWriteStrings[i](out, indent)
		}
	}
	return
}
func (n Named) layout() (l layoutInfo, writeString func(out *strings.Builder, indent []byte)) {
	l.Tokens = []int{1}
	writeString = func(out *strings.Builder, _ []byte) {
		if n.LeftMarker {
			out.WriteString(":")
		}
		if n.Name != "" {
			for _, b := range []byte(n.Name) {
				if byteIsSpecial(b) {
					out.WriteString("\\")
				}
				out.WriteByte(b)
			}
		} else if out.WriteString("_"); true {
		}
		if n.RightMarker {
			out.WriteString(":")
		}
	}
	return
}
func (ISyntax) layout() (l layoutInfo, writeString func(out *strings.Builder, indent []byte)) {
	l.Tokens = []int{2}
	writeString = func(out *strings.Builder, _ []byte) { out.WriteString("()") }
	return
}
func (q Quote) layout() (l layoutInfo, writeString func(out *strings.Builder, indent []byte)) {
	subL, subWriteString := subLayout(q.Syntax, 0)
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
	return
}

func (s List) String() string        { return toString(s) }
func (s EmptyLine) String() string   { return toString(s) }
func (s Application) String() string { return toString(s) }
func (s Composition) String() string { return toString(s) }
func (s Named) String() string       { return toString(s) }
func (s ISyntax) String() string     { return toString(s) }
func (s Quote) String() string       { return toString(s) }

func toString(s Syntax) string {
	_, writeString := subLayout(s, ListPrecedence+1)
	var b strings.Builder
	writeString(&b, nil)
	return b.String()
}

func TopString(s Syntax) string {
	_, writeString := s.(Application).Arg.(Quote).Syntax.layout()
	var b strings.Builder
	writeString(&b, nil)
	return b.String()
}

func EscapeFieldName(fieldName string) string {
	return Named{Name: fieldName}.String()
}
