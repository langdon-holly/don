package rel

import (
	. "don/junctive"
	"don/syntax"
)

func Junction(junctive Junctive, juncts []Rel) Rel {
	jc := JunctionRel{
		Junctive: junctive,
		Juncts:   juncts,
		T:        AnyTypePtr(),
	}
	for _, junct := range juncts {
		UnifyTypePtrs(jc.T, junct.Type())
	}
	return jc
}

type JunctionRel struct {
	Junctive
	Juncts []Rel
	T      *TypePtr
}

func (jc JunctionRel) Type() *TypePtr { return jc.T }

func (jc JunctionRel) Copy(mapping map[*TypePtr]*TypePtr) Rel {
	juncts := make([]Rel, len(jc.Juncts))
	for i, junct := range jc.Juncts {
		juncts[i] = junct.Copy(mapping)
	}
	jc.Juncts = juncts
	jc.T = CopyTypePtr(jc.T, mapping)
	return jc
}
func (jc JunctionRel) Convert() Rel {
	for i, junct := range jc.Juncts {
		jc.Juncts[i] = junct.Convert()
	}
	jc.T = ConvertTypePtr(jc.T)
	return jc
}
func JunctionSyntax(junctive Junctive, juncts [][]syntax.Word) Syntax {
	if len(juncts) == 0 {
		if junctive == ConJunctive {
			return NameSyntax("true")
		} else {
			return NameSyntax("false")
		}
	} else {
		operator := syntax.Word{
			Strings:  []string{"", ""},
			Specials: []syntax.WordSpecial{syntax.WordSpecialJunction(junctive)},
		}

		junctCompositions := append([][]syntax.Word{nil}, juncts...)
		operators := make([]syntax.Word, len(juncts))
		for i := range operators {
			operators[i] = operator
		}
		return SyntaxWords(syntax.Words{Compositions: junctCompositions, Operators: operators})
	}
}
func (jc JunctionRel) Syntax() Syntax {
	junctWordses := make([][]syntax.Word, len(jc.Juncts))
	for i, junct := range jc.Juncts {
		junctWordses[i] = junct.Syntax().Composition()
	}
	return JunctionSyntax(jc.Junctive, junctWordses)
}
func (jc JunctionRel) String() string { return jc.Syntax().String() }
