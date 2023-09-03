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

type WordSpecialDelimited struct {
	LeftDelim, RightDelim MaybeDelim
	Words
}
type WordSpecialJunct Junctive
type WordSpecialCommentMarker struct{}

func (wsd WordSpecialDelimited) layout() (intoMultiline bool, ws writeString) {
	intoMultiline = true
	wordsIntoMultiline, wordsWriteString := wsd.Words.layout()
	if wordsIntoMultiline {
		ws = func(out *strings.Builder, indent []byte) {
			subIndent := append(indent, tab)
			wsd.LeftDelim.writeLeftString(out)
			//out.WriteByte(lf)
			//out.Write(subIndent)
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

// len(Strings) == len(Specials) + 1
// Strings != []string{""}
type Word struct {
	Strings  []string
	Specials []WordSpecial
}

func (w Word) String() string {
	return Words{Compositions: [][]Word{{w}}}.String()
}

// Each []Word != nil
type Words struct{ Compositions [][]Word }

func interCompositionWriteNewline(out *strings.Builder, indent []byte) {
	out.WriteByte(lf)
	out.Write(indent)
}
func interCompositionWriteSpace(out *strings.Builder, indent []byte) {
	out.WriteByte(space)
	out.WriteByte(pipe)
	//out.WriteByte(space)
}
func noopWriteString(out *strings.Builder, indent []byte) {}
func (words Words) layout() (intoMultiline bool, ws writeString) {
	if 0 < len(words.Compositions) {
		specialWriteStringseses := make([][][]writeString, len(words.Compositions)) /* Each [][]writeString non-nil */
		for i, composition := range words.Compositions {
			specialWriteStringses := make([][]writeString, len(composition))
			// 0 < len(specialWriteStringses) == len(composition)
			for j, word := range composition {
				specialWriteStrings := make([]writeString, len(word.Specials))
				// len(specialWriteStrings) == len(word.Specials)
				for k, special := range word.Specials {
					var specialIntoMultiline bool
					specialIntoMultiline, specialWriteStrings[k] = special.layout()
					intoMultiline = intoMultiline || specialIntoMultiline
				}
				specialWriteStringses[j] = specialWriteStrings
			}
			// In specialWriteStringses,
			// 	each len([]writeString) == len(the corresponding Word.Specials in composition)
			specialWriteStringseses[i] = specialWriteStringses
		}
		// In specialWriteStringseses, each len([]writeString)
		// 	== len(the corresponding Word.Specials in words.Compositions)

		var interCompositionWriteString writeString
		if intoMultiline {
			interCompositionWriteString = interCompositionWriteNewline
		} else {
			interCompositionWriteString = interCompositionWriteSpace
		}

		ws = func(out *strings.Builder, indent []byte) {
			for i := 0; ; {
				// 0 < len(words.Compositions)
				// 0 < len(specialWriteStringseses) == len(words.Compositions)
				composition := words.Compositions[i]
				specialWriteStringses := specialWriteStringseses[i] /* non-nil */
				// In specialWriteStringses,
				// 	each len([]writeString) == len(the corresponding Word.Specials in composition)
				for j := 0; ; {
					// 0 < len(composition) (by def.)
					// 0 < len(specialWriteStringses)
					theStrings := composition[j].Strings
					specialWriteStrings := specialWriteStringses[j]
					// len(specialWriteStringses[j] == len(composition[j].Specials), and
					// len(composition[j].Strings) == len(composition[j].Specials) + 1 (by def.), so
					// 	len(theStrings) == len(specialWriteStrings) + 1
					for k := 0; ; k++ {
						out.WriteString(theStrings[k])
						if k >= len(specialWriteStrings) {
							break
						}
						specialWriteStrings[k](out, indent)
					}
					j++
					if j >= len(composition) {
						break
					}
					out.WriteByte(space)
				}
				i++
				if i >= len(words.Compositions) {
					break
				}
				interCompositionWriteString(out, indent)
			}
		}
	} else {
		ws = noopWriteString
	}
	return
}

func (words Words) String() string {
	_, ws := words.layout()
	var b strings.Builder
	ws(&b, nil)
	return b.String()
}
