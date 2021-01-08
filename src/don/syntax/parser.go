package syntax

import (
	"fmt"
	"io"
	"strings"
)

func readByte(r io.Reader) (b byte, ok bool) {
	var bs [1]byte
	for {
		if n, err := r.Read(bs[:]); n == 1 {
			return bs[0], true
		} else if err != nil {
			return 0, false
		}
	}
}

type token struct {
	Bp     bool
	B      byte   /* for Bp */
	Syntax Syntax /* for !Bp */
}

func (t token) IsByte(b byte) bool { return t.Bp && t.B == b }

// Used in spirit
// Discard after bad != nil or Done is called
type parser interface {
	Next(e token) (bad []string)         /* mutates */
	Done() (syntax Syntax, bad []string) /* mutates; syntax for bad != nil */
}

type composition []Syntax

// impl parser
func (state *composition) Next(e token) (bad []string) {
	if *state = append(*state, e.Syntax); e.Bp {
		bad = []string{"Stray byte: '" + string([]byte{e.B}) + "'"}
	}
	return
}
func (state *composition) Done() (syntax Syntax, bad []string) {
	syntax = Syntax{Tag: CompositionSyntaxTag, Children: *state}
	if len(syntax.Children) == 1 {
		syntax = syntax.Children[0]
	}
	return
}

type application struct {
	Sub  composition
	Comp bool
	Com  Syntax
}

func (state *application) maybeInParam(bad *[]string /* mutated */) {
	if state.Comp {
		*bad = append(*bad, "in application parameter")
	}
}

// impl parser
func (state *application) Next(e token) (bad []string) {
	if e.IsByte(bang) {
		subS := &state.Com
		if state.Comp {
			state.Com =
				Syntax{Tag: ApplicationSyntaxTag, Children: []Syntax{state.Com, {}}}
			subS = &state.Com.Children[1]
		}
		*subS, bad = state.Sub.Done()
		if bad == nil && subS.Tag == CompositionSyntaxTag && len(subS.Children) == 0 {
			bad = []string{"Nothing"}
		}
		if bad != nil {
			if state.Comp {
				bad = append(bad, "in application parameter")
			}
			bad = append(bad, "in application computer")
		}
		state.Sub = application{}.Sub
		state.Comp = true
	} else if bad = state.Sub.Next(e); bad != nil && state.Comp {
		bad = append(bad, "in application parameter")
	}
	return
}
func (state *application) Done() (syntax Syntax, bad []string) {
	if state.Comp {
		syntax = Syntax{Tag: ApplicationSyntaxTag, Children: []Syntax{state.Com, {}}}
		subS := &syntax.Children[1]
		if *subS, bad = state.Sub.Done(); bad != nil {
			bad = append(bad, "in application parameter")
		} else if subS.Tag == CompositionSyntaxTag && len(subS.Children) == 0 {
			bad = []string{"Nothing", "in application parameter"}
		}
	} else if syntax, bad = state.Sub.Done(); true {
	}
	return
}

// Children for !Listp
type list struct {
	Sub           application
	Listp         bool
	Midfactor     bool
	EmptyLine     bool
	EmptyLineNext bool
	Children      []Syntax
}

// impl parser
func (state *list) Next(e token) (bad []string) {
	if e.IsByte(asterisk) {
		if factorS, factorBad := state.Sub.Done(); factorBad != nil {
			bad = append(factorBad, "in list")
		} else if factorS.Tag != CompositionSyntaxTag || len(factorS.Children) > 0 {
			if state.EmptyLine && len(state.Children) > 0 {
				state.Children = append(state.Children, Syntax{Tag: EmptyLineSyntaxTag})
			}
			state.Children = append(state.Children, factorS)
		} else if state.Listp {
			bad = []string{"Nothing", "in list"}
		}
		state.Sub = list{}.Sub
		state.Listp = true
		state.Midfactor = false
		state.EmptyLine = state.EmptyLineNext
		state.EmptyLineNext = false
	} else if e.Bp || e.Syntax.Tag != EmptyLineSyntaxTag {
		state.Midfactor = true
		state.EmptyLineNext = false
		if bad = state.Sub.Next(e); bad != nil && state.Listp {
			bad = append(bad, "in list")
		}
	} else if state.Midfactor {
		state.EmptyLineNext = true
	} else if state.EmptyLine = true; true {
	}
	return
}
func (state *list) Done() (syntax Syntax, bad []string) {
	if syntax, bad = state.Sub.Done(); !state.Listp {
	} else if syntax.Tag != CompositionSyntaxTag || len(syntax.Children) > 0 {
		if state.EmptyLine && len(state.Children) > 0 {
			state.Children = append(state.Children, Syntax{Tag: EmptyLineSyntaxTag})
		}
		syntax = Syntax{Tag: ListSyntaxTag, Children: append(state.Children, syntax)}
	} else {
		syntax = Syntax{Tag: ListSyntaxTag, Children: state.Children}
	}
	return
}

type circumfix struct {
	Subs   []list
	Quotes []bool
	Sub    list
}

// impl parser
func (state *circumfix) Next(e token) (bad []string) {
	var subS Syntax
	if e.IsByte(leftParen) {
		state.Subs = append(state.Subs, state.Sub)
		state.Sub = circumfix{}.Sub
		state.Quotes = append(state.Quotes, false)
	} else if e.IsByte(rightParen) {
		if len(state.Subs) == 0 {
			bad = []string{"Not enough left-parens"}
		} else if state.Quotes[len(state.Quotes)-1] {
			bad = []string{"Brace starts, paren ends"}
		} else if subS, bad = state.Sub.Done(); bad == nil {
			if subS.Tag == CompositionSyntaxTag && len(subS.Children) == 0 {
				subS = Syntax{Tag: ISyntaxTag}
			}
			state.Sub = state.Subs[len(state.Subs)-1]
			state.Subs = state.Subs[:len(state.Subs)-1]
			state.Quotes = state.Quotes[:len(state.Quotes)-1]
			bad = state.Sub.Next(token{Syntax: subS})
		}
	} else if e.IsByte(leftBrace) {
		state.Subs = append(state.Subs, state.Sub)
		state.Sub = circumfix{}.Sub
		state.Quotes = append(state.Quotes, true)
	} else if e.IsByte(rightBrace) {
		if len(state.Subs) == 0 {
			bad = []string{"Not enough left-braces"}
		} else if !state.Quotes[len(state.Quotes)-1] {
			bad = []string{"Paren starts, brace ends"}
		} else if subS, bad = state.Sub.Done(); bad == nil {
			if subS.Tag != CompositionSyntaxTag || len(subS.Children) > 0 {
				state.Sub = state.Subs[len(state.Subs)-1]
				state.Subs = state.Subs[:len(state.Subs)-1]
				state.Quotes = state.Quotes[:len(state.Quotes)-1]
				bad = state.Sub.Next(
					token{Syntax: Syntax{Tag: QuotationSyntaxTag, Children: []Syntax{subS}}})
			} else if bad = []string{"Nothing"}; true {
			}
		}
	} else {
		bad = state.Sub.Next(e)
	}
	if bad != nil {
		for i := len(state.Quotes) - 1; i >= 0; i-- {
			if state.Quotes[i] {
				bad = append(bad, "in quotation")
			} else if bad = append(bad, "in parentheses"); true {
			}
		}
	}
	return
}
func (state *circumfix) Done() (syntax Syntax, bad []string) {
	if len(state.Subs) > 0 {
		bad = []string{"Not enough right delimiters"}
		for i := len(state.Quotes) - 1; i >= 0; i-- {
			if state.Quotes[i] {
				bad = append(bad, "in quotation")
			} else if bad = append(bad, "in parentheses"); true {
			}
		}
	} else if syntax, bad = state.Sub.Done(); true {
		s := syntax
		if bad == nil && s.Tag == CompositionSyntaxTag && len(s.Children) == 0 {
			bad = []string{"Nothing"}
		}
	}
	return
}

type name struct {
	Escaped      bool
	NameProgress int /* 0: ready 1: left marked, 2: in name, 3: not ready */
	LeftMarker   bool
	Name         strings.Builder
	Sub          circumfix
	NonemptyLine bool
}

func (state *name) Next(b byte) (bad []string) {
	nonemptyLineNext := true
	if isSpecial := byteIsSpecial(b); !isSpecial && state.NameProgress == 3 {
		bad = []string{"Unseparated names"}
	} else if state.Escaped || !isSpecial {
		state.Escaped = false
		state.NameProgress = 2
		state.Name.WriteByte(b)
	} else if b == underscore {
		if state.NameProgress == 3 {
			bad = []string{"Unseparated names"}
		} else if state.NameProgress = 2; true {
		}
	} else if b == backslash {
		if state.NameProgress == 3 {
			bad = []string{"Unseparated names"}
		} else if state.Escaped = true; true {
		}
	} else if b == colon {
		if state.NameProgress == 0 {
			state.LeftMarker = true
		} else if state.NameProgress == 1 {
			bad = []string{"Double left colon"}
		} else if state.NameProgress == 2 {
			bad = state.Sub.Next(token{Syntax: Syntax{
				Tag:         NameSyntaxTag,
				LeftMarker:  state.LeftMarker,
				RightMarker: true,
				Name:        state.Name.String()}})
			state.LeftMarker = false
			state.NameProgress = 3
			state.Name = strings.Builder{}
		} else if bad = []string{"Unseparated names"}; true {
		}
	} else {
		if state.NameProgress == 1 {
			bad = []string{"Lone colon"}
		} else if state.NameProgress == 2 {
			bad = state.Sub.Next(token{Syntax: Syntax{
				Tag:        NameSyntaxTag,
				LeftMarker: state.LeftMarker,
				Name:       state.Name.String()}})
			state.LeftMarker = false
			state.Name = strings.Builder{}
		}
		if state.NameProgress = 0; bad != nil {
		} else if b != tab && b != lf && b != space {
			bad = state.Sub.Next(token{Bp: true, B: b})
		} else if nonemptyLineNext = state.NonemptyLine && b != lf; false {
		} else if !state.NonemptyLine && b == lf {
			bad = state.Sub.Next(token{Syntax: Syntax{Tag: EmptyLineSyntaxTag}})
		}
	}
	state.NonemptyLine = nonemptyLineNext
	return
}
func (state *name) Done() (syntax Syntax, bad []string) {
	if state.Escaped {
		bad = []string{"Unending escape"}
	} else if state.NameProgress == 1 {
		bad = []string{"Lone colon"}
	} else if state.NameProgress == 2 {
		bad = state.Sub.Next(token{Syntax: Syntax{
			Tag:        NameSyntaxTag,
			LeftMarker: state.LeftMarker,
			Name:       state.Name.String()}})
	}
	if bad == nil {
		syntax, bad = state.Sub.Done()
		syntax = Syntax{
			Tag: ApplicationSyntaxTag,
			Children: []Syntax{
				{Tag: NameSyntaxTag, Name: "context"},
				{Tag: QuotationSyntaxTag, Children: []Syntax{syntax}}}}
	}
	return
}

func doBad(bad []string) {
	if bad != nil {
		for _, s := range bad {
			fmt.Println(s)
		}
		panic("Syntax error")
	}
}

func ParseTop(inReader io.Reader) Syntax {
	var topParser name
	for b, ok := readByte(inReader); ok; b, ok = readByte(inReader) {
		doBad(topParser.Next(b))
	}
	s, bad := topParser.Done()
	doBad(bad)
	return s
}
