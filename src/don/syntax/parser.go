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

type macroCallOrSandwich struct {
	SubName                name
	SubMacroCallOrSandwich *macroCallOrSandwich
	SubMacroCallP          bool
	LHS                    Syntax
}

// impl parser
func (state *macroCallOrSandwich) Next(e token) (bad []string) {
	if state.SubMacroCallOrSandwich != nil {
		if bad = state.SubMacroCallOrSandwich.Next(e); bad == nil {
		} else if state.SubMacroCallP {
			bad = append(bad, "in parameter to macro")
		} else {
			bad = append(bad, "in sandwich liner")
		}
	} else if e.IsByte(bang) {
		state.LHS, bad = state.SubName.Done()
		if bad != nil {
			bad = append(bad, "in macro name")
		} else {
			state.SubMacroCallOrSandwich = new(macroCallOrSandwich)
			state.SubMacroCallP = true
		}
	} else if e.IsByte(hyphen) {
		state.LHS, bad = state.SubName.Done()
		if bad != nil {
			bad = append(bad, "in sandwich bread")
		} else {
			state.SubMacroCallOrSandwich = new(macroCallOrSandwich)
		}
	} else if bad = state.SubName.Next(e); true {
	}
	return
}
func (state *macroCallOrSandwich) Done() (syntax Syntax, bad []string) {
	if state.SubMacroCallOrSandwich == nil {
		syntax, bad = state.SubName.Done()
	} else if state.SubMacroCallP {
		syntax = Syntax{Tag: MCallSyntaxTag, Children: []Syntax{state.LHS, {}}}
		if syntax.Children[1], bad = state.SubMacroCallOrSandwich.Done(); bad != nil {
			bad = append(bad, "in parameter to macro")
		}
	} else {
		syntax = Syntax{Tag: SandwichSyntaxTag, Children: []Syntax{state.LHS, {}}}
		if syntax.Children[1], bad = state.SubMacroCallOrSandwich.Done(); bad != nil {
			bad = append(bad, "in sandwich liner")
		}
	}
	return
}

type spaced struct {
	Midfactor bool
	Sub       macroCallOrSandwich
	Children  []Syntax
}

// impl parser
func (state *spaced) Next(e token) (bad []string) {
	if !e.IsByte(space) {
		state.Midfactor = true
		bad = state.Sub.Next(e)
	} else if state.Midfactor {
		var subS Syntax
		if subS, bad = state.Sub.Done(); bad != nil {
			bad = append(bad, "in spaced")
			return
		}
		state.Midfactor = false
		state.Sub = spaced{}.Sub
		state.Children = append(state.Children, subS)
	} else {
		bad = []string{"Too much space"}
	}
	return
}
func (state *spaced) Done() (syntax Syntax, bad []string) {
	if state.Midfactor {
		var subS Syntax
		if subS, bad = state.Sub.Done(); bad != nil {
			bad = append(bad, "in spaced")
			return
		}
		syntax.Tag = SpacedSyntaxTag
		syntax.Children = append(state.Children, subS)
	} else {
		bad = []string{"Spaced didn't end with factor"}
	}
	return
}

// Only Sub (and Passthrough) for Passthrough
type list struct {
	InitLF      bool
	Passthrough bool
	Midline     bool
	Commented   bool
	Sub         spaced
	Children    []Syntax
}

// impl parser
func (state *list) Next(e token) (bad []string) {
	if state.Passthrough {
		if e.IsByte(lf) ||
			e.IsByte(tab) ||
			e.IsByte(hash) {
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
			} else if !state.Commented {
				state.Children = append(state.Children, subS)
			}
			state.Midline = false
			state.Sub = list{}.Sub
		}
		state.Commented = false
	} else if e.IsByte(tab) {
		if state.Midline {
			bad = []string{"Unindenting tab"}
		}
	} else if !e.IsByte(hash) {
		bad = state.Sub.Next(e)
		state.Midline = true
	} else if state.Midline {
		bad = []string{"End-of-line comment"}
	} else {
		state.Midline = true
		state.Commented = true
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

type parens struct {
	Subs []list
	Sub  list
}

// impl parser
func (state *parens) Next(e token) (bad []string) {
	var subS Syntax
	if e.IsByte(leftParen) {
		state.Subs = append(state.Subs, state.Sub)
		state.Sub = parens{}.Sub
	} else if !e.IsByte(rightParen) {
		bad = state.Sub.Next(e)
	} else if len(state.Subs) == 0 {
		bad = []string{"Not enough left-parens"}
	} else if subS, bad = state.Sub.Done(); bad == nil {
		state.Sub = state.Subs[len(state.Subs)-1]
		state.Subs = state.Subs[:len(state.Subs)-1]
		bad = state.Sub.Next(token{Tag: syntaxTokenTag, Syntax: subS})
	}
	return
}
func (state *parens) Done() (syntax Syntax, bad []string) {
	if len(state.Subs) > 0 {
		bad = []string{"Not enough right-parens"}
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
	var parensParser parens
	for {
		var t token
		if escaped, t = escapeNext(escaped, b); !escaped {
			doBad(parensParser.Next(t))
		}
		if b, ok = readByte(inReader); !ok {
			break
		}
	}
	doBad(escapeDone(escaped))
	s, bad := parensParser.Done()
	doBad(bad)
	if s.Tag != ListSyntaxTag {
		panic("Unreachable")
	} else if len(s.Children) != 1 {
		panic("Not one top-level child")
	}
	return s.Children[0]
}
