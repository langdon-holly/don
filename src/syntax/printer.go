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

type writeString func(out *strings.Builder, indent []byte)

func (Disjunction) precedence() int { return DisjunctionPrecedence }
func (Conjunction) precedence() int { return ConjunctionPrecedence }
func (EmptyLine) precedence() int   { return EmptyLinePrecedence }
func (Application) precedence() int { return LeftAssociativePrecedence }
func (Bind) precedence() int        { return LeftAssociativePrecedence }
func (Composition) precedence() int { return CompositionPrecedence }
func (Named) precedence() int       { return NamedPrecedence }
func (ISyntax) precedence() int     { return ISyntaxPrecedence }
func (Quote) precedence() int       { return QuotePrecedence }

func subLayout(s Syntax, precedence int) (l layoutInfo, ws writeString) {
	if subL, subWriteString := s.layout(); s.precedence() >= precedence {
		l, ws = subL, subWriteString
	} else if len(subL.Tokens) == 2 ||
		subL.Circumfix ||
		subL.Top()+1 > MAX_TOKENS_PER_LINE {
		l = layoutInfo{Circumfix: true, Tokens: []int{1, 1}}
		ws = func(out *strings.Builder, indent []byte) {
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
		ws = func(out *strings.Builder, indent []byte) {
			out.WriteString("(")
			subWriteString(out, indent)
			out.WriteString(")")
		}
	}
	return
}
func layoutListAssociative(items []Syntax, opToken string) (l layoutInfo, ws writeString) {
	if len(items) == 0 {
		l.Tokens = []int{1}
		ws = func(out *strings.Builder, _ []byte) { out.WriteString(opToken) }
	} else if len(items) == 1 {
		var itemWriteString func(*strings.Builder, []byte)
		l, itemWriteString = subLayout(items[0], LeftAssociativePrecedence+1)
		l.Tokens[0]++
		ws = func(out *strings.Builder, indent []byte) {
			out.WriteString(opToken + " ")
			itemWriteString(out, indent)
		}
	} else if itemEmptyLines := make([]bool, len(items)); true {
		itemLs := make([]layoutInfo, len(items))
		itemWriteStrings := make([]func(*strings.Builder, []byte), len(items))
		multiline := false
		tokens := 0
		for i, item := range items {
			_, itemEmptyLines[i] = item.(EmptyLine)
			itemLs[i], itemWriteStrings[i] = subLayout(item, LeftAssociativePrecedence+1)
			multiline = multiline || len(itemLs[i].Tokens) == 2
			l.Circumfix = l.Circumfix || itemLs[i].Circumfix
			tokens += itemLs[i].Top() + 1
		}
		tokens--
		if multiline || tokens > MAX_TOKENS_PER_LINE {
			l.Tokens = []int{
				itemLs[0].Top() + 1,
				layoutInfo{Tokens: []int{1}}.Compose(0, itemLs[len(itemLs)-1]).Bot(),
			}
			ws = func(out *strings.Builder, indent []byte) {
				out.WriteString(opToken + " ")
				itemWriteStrings[0](out, indent)
				for i := 1; i < len(itemWriteStrings); i++ {
					if out.WriteString("\n"); !itemEmptyLines[i] {
						out.Write(indent)
						out.WriteString(opToken + " ")
						itemWriteStrings[i](out, indent)
					}
				}
			}
		} else if l.Tokens = []int{tokens}; true {
			ws = func(out *strings.Builder, indent []byte) {
				itemWriteStrings[0](out, indent)
				for i := 1; i < len(itemWriteStrings); i++ {
					out.WriteString(" " + opToken + " ")
					itemWriteStrings[i](out, indent)
				}
			}
		}
	}
	return
}
func (d Disjunction) layout() (layoutInfo, writeString) {
	return layoutListAssociative(d.Disjuncts, ";")
}
func (c Conjunction) layout() (layoutInfo, writeString) {
	return layoutListAssociative(c.Conjuncts, ",")
}
func (EmptyLine) layout() (l layoutInfo, ws writeString) {
	l.Tokens = []int{0, 0}
	ws = func(*strings.Builder, []byte) { panic("Unreachable") }
	return
}
func layoutLeftAssociative(left, right Syntax, opToken string) (l layoutInfo, ws writeString) {
	leftL, leftWriteString := subLayout(left, LeftAssociativePrecedence)
	rightL, rightWriteString := subLayout(right, LeftAssociativePrecedence+1)
	if len(leftL.Tokens) == 2 || leftL.Bot()+1+rightL.Top() > MAX_TOKENS_PER_LINE {
		l = leftL.Compose(0, layoutInfo{Tokens: []int{0, 1}}).Compose(0, rightL)
		ws = func(out *strings.Builder, indent []byte) {
			leftWriteString(out, indent)
			out.WriteString("\n")
			out.Write(indent)
			out.WriteString(opToken + " ")
			rightWriteString(out, indent)
		}
	} else if l = leftL.Compose(1, rightL); true {
		ws = func(out *strings.Builder, indent []byte) {
			leftWriteString(out, indent)
			out.WriteString(" " + opToken + " ")
			rightWriteString(out, indent)
		}
	}
	return
}
func (a Application) layout() (layoutInfo, writeString) {
	return layoutLeftAssociative(a.Com, a.Arg, "!")
}
func (b Bind) layout() (layoutInfo, writeString) {
	return layoutLeftAssociative(b.Body, b.Var, "?")
}
func (c Composition) layout() (l layoutInfo, ws writeString) {
	l.Tokens = []int{0}
	factorWriteStrings := make([]func(*strings.Builder, []byte), len(c.Factors))
	for i, factor := range c.Factors {
		var factorL layoutInfo
		factorL, factorWriteStrings[i] = subLayout(factor, CompositionPrecedence+1)
		l = l.Compose(0, factorL)
	}
	ws = func(out *strings.Builder, indent []byte) {
		factorWriteStrings[0](out, indent)
		for i := 1; i < len(factorWriteStrings); i++ {
			out.WriteString(" ")
			factorWriteStrings[i](out, indent)
		}
	}
	return
}
func (n Named) layout() (l layoutInfo, ws writeString) {
	l.Tokens = []int{1}
	ws = func(out *strings.Builder, _ []byte) {
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
func (ISyntax) layout() (l layoutInfo, ws writeString) {
	l.Circumfix = true
	l.Tokens = []int{2}
	ws = func(out *strings.Builder, _ []byte) { out.WriteString("()") }
	return
}
func (q Quote) layout() (l layoutInfo, ws writeString) {
	subL, subWriteString := subLayout(q.Syntax, 0)
	if len(subL.Tokens) == 2 ||
		subL.Circumfix ||
		subL.Top()+1 > MAX_TOKENS_PER_LINE {
		l = layoutInfo{Circumfix: true, Tokens: []int{1, 1}}
		ws = func(out *strings.Builder, indent []byte) {
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
		ws = func(out *strings.Builder, indent []byte) {
			out.WriteString("{")
			subWriteString(out, indent)
			out.WriteString("}")
		}
	}
	return
}

func (s Disjunction) String() string { return toString(s) }
func (s Conjunction) String() string { return toString(s) }
func (s EmptyLine) String() string   { return toString(s) }
func (s Application) String() string { return toString(s) }
func (s Bind) String() string        { return toString(s) }
func (s Composition) String() string { return toString(s) }
func (s Named) String() string       { return toString(s) }
func (s ISyntax) String() string     { return toString(s) }
func (s Quote) String() string       { return toString(s) }

func toString(s Syntax) string {
	_, ws := subLayout(s, LeftAssociativePrecedence+1)
	var b strings.Builder
	ws(&b, nil)
	return b.String()
}

func TopString(s Syntax) string {
	_, ws := s.layout()
	var b strings.Builder
	ws(&b, nil)
	return b.String()
}

func EscapeFieldName(fieldName string) string {
	return Named{Name: fieldName}.String()
}
