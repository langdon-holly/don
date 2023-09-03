package rel

import "don/syntax"

type Syntax interface {
	Word() syntax.Word
	Composition() []syntax.Word
	Words() syntax.Words
	String() string
}

type SyntaxWord syntax.Word
type SyntaxComposition []syntax.Word /* non-nil */
type SyntaxWords syntax.Words

func (w SyntaxWord) Word() syntax.Word {
	return syntax.Word(w)
}
func (w SyntaxWord) Composition() []syntax.Word /* non-nil */ {
	return []syntax.Word{syntax.Word(w)}
}
func (w SyntaxWord) Words() syntax.Words {
	return SyntaxComposition(w.Composition()).Words()
}

func (c SyntaxComposition) Composition() []syntax.Word /* non-nil */ {
	return c
}
func (c SyntaxComposition) Words() syntax.Words {
	return syntax.Words{Compositions: [][]syntax.Word{NameSyntax("!").Composition(), c}}
}
func (c SyntaxComposition) Word() syntax.Word {
	return SyntaxWords(c.Words()).Word()
}

func (ws SyntaxWords) Words() syntax.Words { return syntax.Words(ws) }
func (ws SyntaxWords) Word() syntax.Word {
	return syntax.Word{Strings: []string{"", ""}, Specials: []syntax.WordSpecial{syntax.WordSpecialDelimited{
		LeftDelim:  syntax.MaybeDelimParen,
		RightDelim: syntax.MaybeDelimParen,
		Words:      syntax.Words(ws),
	}}}
}
func (ws SyntaxWords) Composition() []syntax.Word /* non-nil */ {
	return SyntaxWord(ws.Word()).Composition()
}

func (w SyntaxWord) String() string        { return w.Words().String() }
func (c SyntaxComposition) String() string { return c.Words().String() }
func (ws SyntaxWords) String() string      { return ws.Words().String() }

func NameSyntax(name string) Syntax {
	return SyntaxWord(syntax.Word{Strings: []string{name}, Specials: nil})
}
