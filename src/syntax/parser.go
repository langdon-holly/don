package syntax

import (
	"io"
	"strings"
)

import . "don/junctive"

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

type wordsParent struct {
	Words     words
	LeftDelim MaybeDelim
}

type words struct {
	Parent *wordsParent

	// len(Compositions) == len(Operators)
	Compositions [][]Word
	Operators    []Word
	Composition  []Word

	// len(Strings) == len(Specials)
	Strings   []string
	Specials  []WordSpecial
	OperatorP bool
	String    strings.Builder
}

type token interface{}

type tokenWS struct{}
type tokenWordSpecial WordSpecial
type tokenLDelim MaybeDelim
type tokenRDelim MaybeDelim

func (state *words) nextSpecial(ws WordSpecial) {
	state.Strings = append(state.Strings, state.String.String())
	state.Specials = append(state.Specials, ws)
	state.String = strings.Builder{}
}
func (state *words) endWord() {
	if len(state.Strings) >= 1 || state.String.Len() >= 1 {
		w := Word{
			Strings:  append(state.Strings, state.String.String()),
			Specials: state.Specials,
		}
		if state.OperatorP {
			state.Compositions = append(state.Compositions, state.Composition)
			state.Operators = append(state.Operators, w)
			state.Composition = nil
		} else {
			state.Composition = append(state.Composition, w)
		}
		state.Strings = nil
		state.Specials = nil
		state.OperatorP = false
		state.String = strings.Builder{}
	}
}
func (state *words) doneSelf() Words {
	state.endWord()
	return Words{
		Compositions: append(state.Compositions, state.Composition),
		Operators:    state.Operators,
	}
}

// Non-nil state.Parent
func (state *words) endDelimitation(rDelim MaybeDelim) {
	selfVal := state.doneSelf()
	state.Parent.Words.nextSpecial(WordSpecialDelimited{
		LeftDelim:  state.Parent.LeftDelim,
		RightDelim: rDelim,
		Words:      selfVal,
	})
	*state = state.Parent.Words
}
func (state *words) Next(b byte) {
	var e token
	switch b {
	case tab:
		e = tokenWS{}
	case lf:
		e = tokenWS{}
	case space:
		e = tokenWS{}
	case bang:
		e = tokenWordSpecial(WordSpecialApplication{})
		state.OperatorP = true
	case hash:
		e = tokenWordSpecial(WordSpecialCommentMarker{})
	case leftParen:
		e = tokenLDelim(MaybeDelimParen)
	case rightParen:
		e = tokenRDelim(MaybeDelimParen)
	case comma:
		e = tokenWordSpecial(WordSpecialJunction(ConJunctive))
		state.OperatorP = true
	case period:
		e = tokenWordSpecial(WordSpecialJunct(ConJunctive))
	case colon:
		e = tokenWordSpecial(WordSpecialJunct(DisJunctive))
	case semicolon:
		e = tokenWordSpecial(WordSpecialJunction(DisJunctive))
		state.OperatorP = true
	case at:
		e = tokenWordSpecial(WordSpecialTuple{})
	case leftBrace:
		e = tokenLDelim(MaybeDelimBrace)
	case rightBrace:
		e = tokenRDelim(MaybeDelimBrace)
	}
	switch eVal := e.(type) {
	case tokenWS:
		state.endWord()
	case tokenWordSpecial:
		state.nextSpecial(eVal)
	case tokenLDelim:
		*state = words{Parent: &wordsParent{Words: *state, LeftDelim: MaybeDelim(eVal)}}
	case tokenRDelim:
		if state.Parent == nil {
			state.Parent = &wordsParent{}
		}
		state.endDelimitation(MaybeDelim(eVal))
	case nil:
		state.String.WriteByte(b)
	}
}
func (state words) Done() Words {
	for {
		if state.Parent == nil {
			return state.doneSelf()
		}
		state.endDelimitation(MaybeDelimNone)
	}
}

func Parse(inReader io.Reader) Words {
	var topWords words
	for {
		b, ok := readByte(inReader)
		if !ok {
			break
		}
		topWords.Next(b)
	}
	return topWords.Done()
}
