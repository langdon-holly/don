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
func (state composition) Done() (syntax Syntax, bad []string) {
	if syntax = (Composition{state}); len(state) == 1 {
		syntax = state[0]
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
		var subS Syntax
		subS, bad = state.Sub.Done()
		if state.Comp {
			state.Com = Application{Com: state.Com, Arg: subS}
		} else if state.Com = subS; true {
		}
		if bad == nil && isEmptyComposition(subS) {
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
	var subS Syntax
	subS, bad = state.Sub.Done()
	if state.Comp {
		syntax = Application{Com: state.Com, Arg: subS}
		if bad != nil {
			bad = append(bad, "in application parameter")
		} else if isEmptyComposition(subS) {
			bad = []string{"Nothing", "in application parameter"}
		}
	} else if syntax, bad = state.Sub.Done(); true {
	}
	return
}

// Factors for !Listp
type list struct {
	Sub           application
	Listp         bool
	Midfactor     bool
	EmptyLine     bool
	EmptyLineNext bool
	Factors       []Syntax
}

// impl parser
func (state *list) Next(e token) (bad []string) {
	if e.IsByte(asterisk) {
		if factorS, factorBad := state.Sub.Done(); factorBad != nil {
			bad = append(factorBad, "in list")
		} else if !isEmptyComposition(factorS) {
			if state.EmptyLine && len(state.Factors) > 0 {
				state.Factors = append(state.Factors, EmptyLine{})
			}
			state.Factors = append(state.Factors, factorS)
		} else if state.Listp {
			bad = []string{"Nothing", "in list"}
		}
		state.Sub = list{}.Sub
		state.Listp = true
		state.Midfactor = false
		state.EmptyLine = state.EmptyLineNext
		state.EmptyLineNext = false
	} else if _, emptyp := e.Syntax.(EmptyLine); !emptyp {
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
	} else if isEmptyComposition(syntax) {
		syntax = List{state.Factors}
	} else if state.EmptyLine && len(state.Factors) > 0 {
		syntax = List{append(state.Factors, EmptyLine{}, syntax)}
	} else if syntax = (List{append(state.Factors, syntax)}); true {
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
		syntax = Application{Com: Named{Name: "context"}, Arg: Quote{syntax}}
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
