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
	layout() (intoMultiline bool, ws writeString)
}

// Not operator
type WordSpecialDelimited struct {
	LeftDelim, RightDelim MaybeDelim
	Words
}
type WordSpecialJunct Junctive
type WordSpecialCommentMarker struct{}
type WordSpecialTuple struct{}

// Operator
type WordSpecialJunction Junctive
type WordSpecialApplication struct{}

func (wsd WordSpecialDelimited) layout() (intoMultiline bool, ws writeString) {
	intoMultiline = true
	wordsIntoMultiline, wordsWriteString := wsd.Words.layout()
	if wordsIntoMultiline {
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
func (wsj WordSpecialJunct) layout() (intoMultiline bool, ws writeString) {
	ws = func(out *strings.Builder, _ []byte) {
		if Junctive(wsj) == ConJunctive {
			out.WriteByte(period)
		} else if out.WriteByte(colon); true {
		}
	}
	return
}
func (_ WordSpecialCommentMarker) layout() (intoMultiline bool, ws writeString) {
	ws = func(out *strings.Builder, _ []byte) { out.WriteByte(hash) }
	return
}
func (_ WordSpecialTuple) layout() (intoMultiline bool, ws writeString) {
	ws = func(out *strings.Builder, _ []byte) { out.WriteByte(at) }
	return
}

func (wsj WordSpecialJunction) layout() (intoMultiline bool, ws writeString) {
	ws = func(out *strings.Builder, _ []byte) {
		if Junctive(wsj) == ConJunctive {
			out.WriteByte(comma)
		} else if out.WriteByte(semicolon); true {
		}
	}
	return
}
func (_ WordSpecialApplication) layout() (intoMultiline bool, ws writeString) {
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
func wordSliceLayout(wordSlice []Word) (
	intoMultiline bool,
	specialWriteStringses [][]writeString,
) {
	specialWriteStringses = make([][]writeString, len(wordSlice))
	for i, word := range wordSlice {
		specialWriteStrings := make([]writeString, len(word.Specials))
		for j, special := range word.Specials {
			var specialIntoMultiline bool
			specialIntoMultiline, specialWriteStrings[j] = special.layout()
			intoMultiline = intoMultiline || specialIntoMultiline
		}
		specialWriteStringses[i] = specialWriteStrings
	}
	return
}
func writeComposition(
	out *strings.Builder,
	indent []byte,
	wordAlready *bool,
	composition []Word,
	compositionSpecialWriteStringses [][]writeString,
) {
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
func (words Words) layout() (intoMultiline bool, ws writeString) {
	compositionSpecialWriteStringseses := make([][][]writeString, len(words.Compositions))
	for i, composition := range words.Compositions {
		var compositionIntoMultiline bool
		compositionIntoMultiline, compositionSpecialWriteStringseses[i] =
			wordSliceLayout(composition)
		intoMultiline = intoMultiline || compositionIntoMultiline
	}

	operatorsIntoMultiline, operatorSpecialWriteStringses := wordSliceLayout(words.Operators)
	intoMultiline = intoMultiline || operatorsIntoMultiline

	var preOpWriteString writeString
	if intoMultiline {
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
