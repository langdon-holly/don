package syntax

import "strings"

import . "don/junctive"

const MAX_TOKENS_PER_LINE = 16 /* Suggestion */

type writeString = func(out *strings.Builder, indent []byte)

type MaybeDelim int

const (
	MaybeDelimNone = iota
	MaybeDelimParen
	MaybeDelimBrace
)

func (delim MaybeDelim) writeLeftString(out *strings.Builder) {
	switch delim {
	case MaybeDelimNone:
	case MaybeDelimParen:
		out.WriteByte(leftParen)
	case MaybeDelimBrace:
		out.WriteByte(leftBrace)
	default:
		panic("Unreachable")
	}
}
func (delim MaybeDelim) writeRightString(out *strings.Builder) {
	switch delim {
	case MaybeDelimNone:
	case MaybeDelimParen:
		out.WriteByte(rightParen)
	case MaybeDelimBrace:
		out.WriteByte(rightBrace)
	default:
		panic("Unreachable")
	}
}

func (delim MaybeDelim) String() string {
	switch delim {
	case MaybeDelimNone:
		return "none"
	case MaybeDelimParen:
		return "paren"
	case MaybeDelimBrace:
		return "brace"
	default:
		panic("Unreachable")
	}
}

type WordSpecial interface {
	layout() (tokens int, ws writeString)
}

// Not operator
type WordSpecialDelimited struct {
	LeftDelim, RightDelim MaybeDelim
	Words
}
type WordSpecialJunct Junctive
type WordSpecialCommentMarker struct{}

// Operator
type WordSpecialJunction Junctive
type WordSpecialApplication struct{}

func (wsd WordSpecialDelimited) layout() (tokens int, ws writeString) {
	wordsTokens, wordsWriteString := wsd.Words.layout()
	tokens = wordsTokens + 2 /* For simplicity, count MaybeDelimNone as a token */
	if tokens > MAX_TOKENS_PER_LINE {
		ws = func(out *strings.Builder, indent []byte) {
			subIndent := append(indent, tab)
			wsd.LeftDelim.writeLeftString(out)
			out.WriteByte(lf)
			out.Write(subIndent)
			wordsWriteString(out, subIndent)
			out.WriteByte(lf)
			out.Write(indent)
			wsd.RightDelim.writeRightString(out)
		}
	} else {
		ws = func(out *strings.Builder, indent []byte) {
			subIndent := append(indent, tab)
			wsd.LeftDelim.writeLeftString(out)
			wordsWriteString(out, subIndent)
			wsd.RightDelim.writeRightString(out)
		}
	}
	return
}
func (wsj WordSpecialJunct) layout() (tokens int, ws writeString) {
	tokens = 1
	ws = func(out *strings.Builder, _ []byte) {
		if Junctive(wsj) == ConJunctive {
			out.WriteByte(period)
		} else if out.WriteByte(colon); true {
		}
	}
	return
}
func (_ WordSpecialCommentMarker) layout() (tokens int, ws writeString) {
	tokens = 1
	ws = func(out *strings.Builder, _ []byte) { out.WriteByte(hash) }
	return
}

func (wsj WordSpecialJunction) layout() (tokens int, ws writeString) {
	tokens = 1
	ws = func(out *strings.Builder, _ []byte) {
		if Junctive(wsj) == ConJunctive {
			out.WriteByte(comma)
		} else if out.WriteByte(semicolon); true {
		}
	}
	return
}
func (_ WordSpecialApplication) layout() (tokens int, ws writeString) {
	tokens = 1
	ws = func(out *strings.Builder, _ []byte) { out.WriteByte(bang) }
	return
}

// len(Strings) == len(Specials) + 1
// Strings != []string{""}
type Word struct {
	Strings  []string
	Specials []WordSpecial
}

func (w Word) String() string {
	return Words{Compositions: [][]Word{{w}}}.String()
}

// len(Compositions) == len(Operators) + 1
type Words struct {
	Compositions [][]Word /* Each word has no operator byte */
	Operators    []Word   /* Each word has an operator byte */
}

func wordWriteString(
	out *strings.Builder,
	indent []byte,
	theStrings []string,
	specialWriteStrings []writeString,
) {
	for j := 0; ; j++ {
		out.WriteString(theStrings[j])

		if j >= len(specialWriteStrings) {
			break
		}

		specialWriteStrings[j](out, indent)
	}
}
func wordSliceLayout(tokens *int, wordSlice []Word) [][]writeString {
	specialWriteStringses := make([][]writeString, len(wordSlice))
	for i, word := range wordSlice {
		specialWriteStrings := make([]writeString, len(word.Specials))
		for j := 0; ; j++ {
			if word.Strings[j] != "" {
				*tokens++
			}

			if j >= len(word.Specials) {
				break
			}

			var specialTokens int
			specialTokens, specialWriteStrings[j] = word.Specials[j].layout()
			*tokens += specialTokens
		}
		specialWriteStringses[i] = specialWriteStrings
	}
	return specialWriteStringses
}
func writeComposition(out *strings.Builder, indent []byte, wordAlready *bool, composition []Word, compositionSpecialWriteStringses [][]writeString) {
	for i, factor := range composition {
		if *wordAlready {
			out.WriteByte(space)
		}
		wordWriteString(out, indent, factor.Strings, compositionSpecialWriteStringses[i])
		*wordAlready = true
	}
}
func preOpWriteNewline(out *strings.Builder, indent []byte) {
	out.WriteByte(lf)
	out.Write(indent)
}
func preOpWriteSpace(out *strings.Builder, indent []byte) {
	out.WriteByte(space)
}
func (words Words) layout() (tokens int, ws writeString) {
	compositionSpecialWriteStringseses := make([][][]writeString, len(words.Compositions))
	for i, composition := range words.Compositions {
		compositionSpecialWriteStringseses[i] = wordSliceLayout(&tokens, composition)
	}
	operatorSpecialWriteStringses := wordSliceLayout(&tokens, words.Operators)
	var preOpWriteString writeString
	if tokens > MAX_TOKENS_PER_LINE {
		preOpWriteString = preOpWriteNewline
	} else {
		preOpWriteString = preOpWriteSpace
	}
	ws = func(out *strings.Builder, indent []byte) {
		wordAlready := false
		for i := 0; ; {
			writeComposition(out, indent, &wordAlready, words.Compositions[i], compositionSpecialWriteStringseses[i])
			if i >= len(words.Operators) {
				break
			}
			if wordAlready {
				preOpWriteString(out, indent)
			}
			wordWriteString(out, indent, words.Operators[i].Strings, operatorSpecialWriteStringses[i])
			wordAlready = true
			i++
		}
	}
	return
}

func (words Words) String() string {
	_, ws := words.layout()
	var b strings.Builder
	ws(&b, nil)
	return b.String()
}
