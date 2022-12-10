package com

import "don/syntax"

type Syntax interface {
	Word() syntax.Word          /* Has no operator byte */
	Composition() []syntax.Word /* Each word has no operator byte */
	Words() syntax.Words
	String() string
}

type SyntaxWord syntax.Word          /* Has no operator byte */
type SyntaxComposition []syntax.Word /* Each word has no operator byte */
type SyntaxWords syntax.Words

func (w SyntaxWord) Word() /* Has no operator byte */ syntax.Word {
	return syntax.Word(w)
}
func (w SyntaxWord) Composition() /* Each word has no operator byte */ []syntax.Word {
	return []syntax.Word{syntax.Word(w)}
}
func (w SyntaxWord) Words() syntax.Words {
	return SyntaxComposition(w.Composition()).Words()
}

func (c SyntaxComposition) Composition() /* Each word has no operator byte */ []syntax.Word {
	return c
}
func (c SyntaxComposition) Words() syntax.Words {
	return syntax.Words{Compositions: [][]syntax.Word{c}, Operators: nil}
}
func (c SyntaxComposition) Word() /* Has no operator byte */ syntax.Word {
	return SyntaxWords(c.Words()).Word()
}

func (ws SyntaxWords) Words() syntax.Words { return syntax.Words(ws) }
func (ws SyntaxWords) Word() /* Has no operator byte */ syntax.Word {
	return syntax.Word{Strings: []string{"", ""}, Specials: []syntax.WordSpecial{syntax.WordSpecialDelimited{
		LeftDelim:  syntax.MaybeDelimParen,
		RightDelim: syntax.MaybeDelimParen,
		Words:      syntax.Words(ws),
	}}}
}
func (ws SyntaxWords) Composition() /* Each word has no operator byte */ []syntax.Word {
	return SyntaxWord(ws.Word()).Composition()
}

func (w SyntaxWord) String() string        { return w.Words().String() }
func (c SyntaxComposition) String() string { return c.Words().String() }
func (ws SyntaxWords) String() string      { return ws.Words().String() }

func NameSyntax(name string) Syntax {
	return SyntaxWord(syntax.Word{Strings: []string{name}, Specials: nil})
}
