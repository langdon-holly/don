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

func isEmptyComposition(s Syntax) bool {
	composition, compositionp := s.(Composition)
	return compositionp && len(composition.Factors) == 0
}

// Discard after bad != nil or Done is called
type parser interface {
	Next(e token) (bad []string)         /* mutates */
	Done() (syntax Syntax, bad []string) /* mutates; syntax for bad == nil */
}

type composition []Syntax

// impl parser
func (state *composition) Next(e token) (bad []string) {
	if *state = append(*state, e.Syntax); e.Bp {
		bad = []string{"Stray byte: '" + string([]byte{e.B}) + "'"}
	}
	return
}
func (state composition) Done() (syntax Syntax, bad []string) {
	if syntax = (Composition{state}); len(state) == 1 {
		syntax = state[0]
	}
	return
}

type leftAssociative struct {
	Sub                   composition
	InApplication, InBind bool /* Exclusive */
	Com                   Syntax
}

// impl parser
func (state *leftAssociative) Next(e token) (bad []string) {
	if banged, questioned := e.IsByte(bang), e.IsByte(question); banged || questioned {
		var subS Syntax
		subS, bad = state.Sub.Done()
		if state.InApplication {
			state.Com = Application{Com: state.Com, Arg: subS}
		} else if state.InBind {
			state.Com = Bind{Body: state.Com, Var: subS}
		} else if state.Com = subS; true {
		}
		if bad == nil && isEmptyComposition(subS) {
			bad = []string{"Nothing"}
		}
		if bad != nil {
			if state.InApplication {
				bad = append(bad, "in application parameter")
			} else if state.InBind {
				bad = append(bad, "in bind variable")
			}
			if banged {
				bad = append(bad, "in application computer")
			} else /* questioned */ {
				bad = append(bad, "in bind body")
			}
		}
		state.Sub = leftAssociative{}.Sub
		state.InApplication = banged
		state.InBind = questioned
	} else if bad = state.Sub.Next(e); bad == nil {
	} else if state.InApplication {
		bad = append(bad, "in application parameter")
	} else if state.InBind {
		bad = append(bad, "in bind variable")
	}
	return
}
func (state *leftAssociative) Done() (syntax Syntax, bad []string) {
	var subS Syntax
	subS, bad = state.Sub.Done()
	if state.InApplication {
		syntax = Application{Com: state.Com, Arg: subS}
		if bad != nil {
			bad = append(bad, "in application parameter")
		} else if isEmptyComposition(subS) {
			bad = []string{"Nothing", "in application parameter"}
		}
	} else if state.InBind {
		syntax = Bind{Body: state.Com, Var: subS}
		if bad != nil {
			bad = append(bad, "in bind variable")
		} else if isEmptyComposition(subS) {
			bad = []string{"Nothing", "in bind variable"}
		}
	} else if syntax, bad = state.Sub.Done(); true {
	}
	return
}

// Items for !Listp
type listAssociative struct {
	Listp         bool
	Miditem       bool
	EmptyLine     bool
	EmptyLineNext bool
	Items         []Syntax
}

// Mutates
func (state *listAssociative) Next(
	sub parser, /* mutated */
	e token,
	opToken byte,
	badMsg string,
) (zeroSub bool, bad []string) {
	if e.IsByte(opToken) {
		if itemS, itemBad := sub.Done(); itemBad != nil {
			bad = append(itemBad, badMsg)
		} else if !isEmptyComposition(itemS) {
			if state.EmptyLine && len(state.Items) > 0 {
				state.Items = append(state.Items, EmptyLine{})
			}
			state.Items = append(state.Items, itemS)
		} else if state.Listp {
			bad = []string{"Nothing", badMsg}
		}
		zeroSub = true
		state.Listp = true
		state.Miditem = false
		state.EmptyLine = state.EmptyLineNext
		state.EmptyLineNext = false
	} else if _, emptyp := e.Syntax.(EmptyLine); !emptyp {
		state.Miditem = true
		state.EmptyLineNext = false
		if bad = sub.Next(e); bad != nil && state.Listp {
			bad = append(bad, badMsg)
		}
	} else if state.Miditem {
		state.EmptyLineNext = true
	} else if state.EmptyLine = true; true {
	}
	return
}

// Mutates
// listp for bad == nil
// items for listp
// syntax for !listp
func (state *listAssociative) Done(sub parser /* mutated */, badMsg string) (
	items []Syntax,
	syntax Syntax,
	listp bool,
	bad []string,
) {
	if syntax, bad = sub.Done(); !state.Listp {
	} else if bad != nil {
		bad = append(bad, badMsg)
	} else if listp = true; isEmptyComposition(syntax) {
		items = state.Items
	} else if state.EmptyLine && len(state.Items) > 0 {
		items = append(state.Items, EmptyLine{}, syntax)
	} else if items = append(state.Items, syntax); true {
	}
	return
}

type conjunction struct {
	Sub leftAssociative
	listAssociative
}

// impl parser
func (state *conjunction) Next(e token) (bad []string) {
	var zeroSub bool
	zeroSub, bad = state.listAssociative.Next(&state.Sub, e, comma, "in conjunction")
	if zeroSub {
		state.Sub = conjunction{}.Sub
	}
	return
}
func (state *conjunction) Done() (syntax Syntax, bad []string) {
	var items []Syntax
	var listp bool
	items, syntax, listp, bad = state.listAssociative.Done(&state.Sub, "in conjunction")
	if listp {
		syntax = Conjunction{items}
	}
	return
}

type disjunction struct {
	Sub conjunction
	listAssociative
}

// impl parser
func (state *disjunction) Next(e token) (bad []string) {
	var zeroSub bool
	zeroSub, bad = state.listAssociative.Next(&state.Sub, e, semicolon, "in disjunction")
	if zeroSub {
		state.Sub = disjunction{}.Sub
	}
	return
}
func (state *disjunction) Done() (syntax Syntax, bad []string) {
	var items []Syntax
	var listp bool
	items, syntax, listp, bad = state.listAssociative.Done(&state.Sub, "in disjunction")
	if listp {
		syntax = Disjunction{items}
	}
	return
}

type circumfix struct {
	Subs   []disjunction
	Quotes []bool
	Sub    disjunction
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
			if isEmptyComposition(subS) {
				subS = ISyntax{}
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
			if !isEmptyComposition(subS) {
				state.Sub = state.Subs[len(state.Subs)-1]
				state.Subs = state.Subs[:len(state.Subs)-1]
				state.Quotes = state.Quotes[:len(state.Quotes)-1]
				bad = state.Sub.Next(token{Syntax: Quote{subS}})
			} else if bad = []string{"Nothing"}; true {
			}
		}
	} else {
		bad = state.Sub.Next(e)
	}
	if bad != nil {
		for i := len(state.Quotes) - 1; i >= 0; i-- {
			if state.Quotes[i] {
				bad = append(bad, "in quote")
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
				bad = append(bad, "in quote")
			} else if bad = append(bad, "in parentheses"); true {
			}
		}
	} else if syntax, bad = state.Sub.Done(); false {
	} else if bad == nil && isEmptyComposition(syntax) {
		bad = []string{"Nothing"}
	}
	return
}

type named struct {
	Escaped      bool
	Progress     int /* 0: ready 1: left marked, 2: in named, 3: not ready */
	LeftMarker   bool
	Name         strings.Builder
	Sub          circumfix
	NonemptyLine bool
}

func (state *named) Next(b byte) (bad []string) {
	nonemptyLineNext := true
	if isSpecial := byteIsSpecial(b); !isSpecial && state.Progress == 3 {
		bad = []string{"Unseparated nameds"}
	} else if state.Escaped || !isSpecial {
		state.Escaped = false
		state.Progress = 2
		state.Name.WriteByte(b)
	} else if b == underscore {
		if state.Progress == 3 {
			bad = []string{"Unseparated nameds"}
		} else if state.Progress = 2; true {
		}
	} else if b == backslash {
		if state.Progress == 3 {
			bad = []string{"Unseparated nameds"}
		} else if state.Escaped = true; true {
		}
	} else if b == colon {
		if state.Progress == 0 {
			state.LeftMarker = true
		} else if state.Progress == 1 {
			bad = []string{"Double left colon"}
		} else if state.Progress == 2 {
			bad = state.Sub.Next(token{Syntax: Named{
				LeftMarker:  state.LeftMarker,
				RightMarker: true,
				Name:        state.Name.String()}})
			state.LeftMarker = false
			state.Progress = 3
			state.Name = strings.Builder{}
		} else if bad = []string{"Unseparated nameds"}; true {
		}
	} else {
		if state.Progress == 1 {
			bad = []string{"Lone colon"}
		} else if state.Progress == 2 {
			bad = state.Sub.Next(token{Syntax: Named{
				LeftMarker: state.LeftMarker,
				Name:       state.Name.String()}})
			state.LeftMarker = false
			state.Name = strings.Builder{}
		}
		if state.Progress = 0; bad != nil {
		} else if b != tab && b != lf && b != space {
			bad = state.Sub.Next(token{Bp: true, B: b})
		} else if nonemptyLineNext = state.NonemptyLine && b != lf; false {
		} else if !state.NonemptyLine && b == lf {
			bad = state.Sub.Next(token{Syntax: EmptyLine{}})
		}
	}
	state.NonemptyLine = nonemptyLineNext
	return
}
func (state *named) Done() (syntax Syntax, bad []string) {
	if state.Escaped {
		bad = []string{"Unending escape"}
	} else if state.Progress == 1 {
		bad = []string{"Lone colon"}
	} else if state.Progress == 2 {
		bad = state.Sub.Next(token{Syntax: Named{
			LeftMarker: state.LeftMarker,
			Name:       state.Name.String()}})
	}
	if bad == nil {
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
	var topParser named
	for b, ok := readByte(inReader); ok; b, ok = readByte(inReader) {
		doBad(topParser.Next(b))
	}
	s, bad := topParser.Done()
	doBad(bad)
	return s
}
