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

type tokenTag int

const (
	byteTokenTag = tokenTag(iota)
	escapedTokenTag
	syntaxTokenTag
)

type token struct {
	Tag tokenTag
	B   byte
	Syntax
}

func (t token) IsByte(b byte) bool { return t.Tag == byteTokenTag && t.B == b }

// t for !nextEscaped
func escapeNext(escaped bool, b byte) (nextEscaped bool, t token) {
	if escaped {
		t = token{Tag: escapedTokenTag, B: b}
	} else if b == backslash {
		nextEscaped = true
	} else {
		t = token{Tag: byteTokenTag, B: b}
	}
	return
}

func escapeDone(escaped bool) (bad []string) {
	if escaped {
		bad = []string{"Unending escape"}
	}
	return
}

// Used in spirit
// Discard after bad != nil or Done is called
type parser interface {
	Next(e token) (bad []string)         /* mutates */
	Done() (syntax Syntax, bad []string) /* mutates; syntax for bad != nil */
}

type name struct {
	Namep, Preparsed        bool
	Nonemptyp               bool
	LeftMarker, RightMarker bool
	B                       strings.Builder
	PPSyntax                Syntax
}

// impl parser
func (state *name) Next(e token) (bad []string) {
	if state.Preparsed || state.Namep && e.Tag == syntaxTokenTag {
		bad = []string{"Subelement in name"}
	} else if e.Tag == syntaxTokenTag {
		state.Preparsed = true
		state.PPSyntax = e.Syntax
	} else if state.RightMarker {
		bad = []string{"Nonterminal right colon"}
	} else if state.Namep = true; e.IsByte(underscore) {
		state.Nonemptyp = true
	} else if !e.IsByte(colon) {
		state.Nonemptyp = true
		state.B.WriteByte(e.B)
	} else if state.Nonemptyp {
		state.RightMarker = true
	} else if state.LeftMarker {
		bad = []string{"Double left colon"}
	} else if state.LeftMarker = true; true {
	}
	return
}
func (state *name) Done() (syntax Syntax, bad []string) {
	if state.Nonemptyp {
		syntax.Tag = NameSyntaxTag
		syntax.LeftMarker = state.LeftMarker
		syntax.RightMarker = state.RightMarker
		syntax.Name = state.B.String()
	} else if state.Preparsed {
		syntax = state.PPSyntax
	} else {
		bad = []string{"Empty name"}
	}
	return
}

type composition struct {
	Midfactor bool
	Sub       name
	Children  []Syntax
}

// impl parser
func (state *composition) Next(e token) (bad []string) {
	if !e.IsByte(space) {
		state.Midfactor = true
		bad = state.Sub.Next(e)
	} else if state.Midfactor {
		var subS Syntax
		if subS, bad = state.Sub.Done(); bad != nil {
			bad = append(bad, "in composition")
			return
		}
		state.Midfactor = false
		state.Sub = composition{}.Sub
		state.Children = append(state.Children, subS)
	} else {
		bad = []string{"Too much space"}
	}
	return
}
func (state *composition) Done() (syntax Syntax, bad []string) {
	if state.Midfactor {
		var subS Syntax
		if subS, bad = state.Sub.Done(); bad != nil {
			bad = append(bad, "in composition")
			return
		}
		syntax.Tag = CompositionSyntaxTag
		syntax.Children = append(state.Children, subS)
		if len(syntax.Children) == 1 {
			syntax = syntax.Children[0]
		}
	} else {
		bad = []string{"Composition didn't end with factor"}
	}
	return
}

type application struct {
	OpProgress int
	Sub        composition
	ComP       bool
	Com        Syntax
}

var spaceToken = token{Tag: byteTokenTag, B: space}

func (state *application) maybeInParam(bad *[]string /* mutated */) {
	if state.ComP {
		*bad = append(*bad, "in application parameter")
	}
}

// impl parser
func (state *application) Next(e token) (bad []string) {
	if state.OpProgress == 0 {
		if e.IsByte(space) {
			state.OpProgress++
		} else if e.IsByte(bang) {
			bad = []string{"Bang not after space"}
			state.maybeInParam(&bad)
		} else if bad = state.Sub.Next(e); bad != nil {
			state.maybeInParam(&bad)
		}
	} else if state.OpProgress == 1 {
		if e.IsByte(space) {
			bad = []string{"Too much space"}
			state.maybeInParam(&bad)
		} else if e.IsByte(bang) {
			state.OpProgress++
			subS := &state.Com
			if state.ComP {
				state.Com = Syntax{Tag: ApplicationSyntaxTag, Children: []Syntax{state.Com, {}}}
				subS = &state.Com.Children[1]
			}
			if *subS, bad = state.Sub.Done(); bad != nil {
				state.maybeInParam(&bad)
				bad = append(bad, "in application computer")
			}
			state.Sub = application{}.Sub
			state.ComP = true
		} else if state.OpProgress = 0; true {
			if bad = state.Sub.Next(spaceToken); bad != nil {
				state.maybeInParam(&bad)
			} else if bad = state.Sub.Next(e); bad != nil {
				state.maybeInParam(&bad)
			}
		}
	} else if state.OpProgress = 0; !e.IsByte(space) {
		bad = []string{"Bang not before space"}
		state.maybeInParam(&bad)
	}
	return
}
func (state *application) Done() (syntax Syntax, bad []string) {
	if state.OpProgress == 2 {
		bad = []string{"Nothing after bang"}
		state.maybeInParam(&bad)
		return
	}
	if state.OpProgress == 1 {
		if bad = state.Sub.Next(spaceToken); bad != nil {
			state.maybeInParam(&bad)
			return
		}
	}
	subS := &syntax
	if state.ComP {
		syntax = Syntax{Tag: ApplicationSyntaxTag, Children: []Syntax{state.Com, {}}}
		subS = &syntax.Children[1]
	}
	if *subS, bad = state.Sub.Done(); bad != nil {
		state.maybeInParam(&bad)
	}
	return
}

// Only Sub (and Passthrough) for Passthrough
type list struct {
	InitLF      bool
	Passthrough bool
	Midline     bool
	Sub         application
	Children    []Syntax
}

// impl parser
func (state *list) Next(e token) (bad []string) {
	if state.Passthrough {
		if e.IsByte(lf) || e.IsByte(tab) {
			bad = []string{"List-specific byte but no initial LF"}
		} else {
			bad = state.Sub.Next(e)
		}
	} else if !state.InitLF {
		if e.IsByte(lf) {
			state.InitLF = true
		} else {
			state.Passthrough = true
			bad = state.Next(e)
		}
	} else if e.IsByte(lf) {
		if state.Midline {
			var subS Syntax
			if subS, bad = state.Sub.Done(); bad != nil {
				bad = append(bad, "at EOL in list")
				return
			} else if state.Children = append(state.Children, subS); true {
			}
			state.Midline = false
			state.Sub = list{}.Sub
		} else {
			state.Children = append(state.Children, Syntax{Tag: EmptyLineSyntaxTag})
		}
	} else if !e.IsByte(tab) {
		bad = state.Sub.Next(e)
		state.Midline = true
	} else if state.Midline {
		bad = []string{"Unindenting tab"}
	}
	return
}
func (state *list) Done() (syntax Syntax, bad []string) {
	if state.Passthrough {
		syntax, bad = state.Sub.Done()
	} else if !state.InitLF {
		bad = []string{"No LF in list"}
	} else if state.Midline {
		bad = []string{"Mid-line end of list"}
	} else {
		syntax.Tag = ListSyntaxTag
		syntax.Children = state.Children
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
			state.Sub = state.Subs[len(state.Subs)-1]
			state.Subs = state.Subs[:len(state.Subs)-1]
			state.Quotes = state.Quotes[:len(state.Quotes)-1]
			bad = state.Sub.Next(token{Tag: syntaxTokenTag, Syntax: subS})
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
			state.Sub = state.Subs[len(state.Subs)-1]
			state.Subs = state.Subs[:len(state.Subs)-1]
			state.Quotes = state.Quotes[:len(state.Quotes)-1]
			bad = state.Sub.Next(token{
				Tag:    syntaxTokenTag,
				Syntax: Syntax{Tag: QuotationSyntaxTag, Children: []Syntax{subS}}})
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
	} else {
		syntax, bad = state.Sub.Done()
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
	b, ok := lf, true
	var escaped bool
	var circumfixParser circumfix
	for {
		var t token
		if escaped, t = escapeNext(escaped, b); !escaped {
			doBad(circumfixParser.Next(t))
		}
		if b, ok = readByte(inReader); !ok {
			break
		}
	}
	doBad(escapeDone(escaped))
	s, bad := circumfixParser.Done()
	doBad(bad)
	return Syntax{
		Tag: ApplicationSyntaxTag,
		Children: []Syntax{
			{Tag: NameSyntaxTag, Name: "context"},
			{Tag: QuotationSyntaxTag, Children: []Syntax{s}}}}
}
