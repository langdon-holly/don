package rel

import (
	. "don/junctive"
	"don/syntax"
)

func Select(junctive Junctive, fieldName string) Rel {
	t := PairTypePtr()
	UnifyTypePtrs(TypePtrAt(junctive, fieldName, GetLeft(t)), GetRight(t))
	return SelectRel{
		FieldName: fieldName,
		Junctive:  junctive,
		T:         t,
	}
}

type SelectRel struct {
	FieldName string
	Junctive
	T *TypePtr
}

func (sc SelectRel) Type() *TypePtr { return sc.T }
func (sc SelectRel) Copy(mapping map[*TypePtr]*TypePtr) Rel {
	sc.T = CopyTypePtr(sc.T, mapping)
	return sc
}
func (sc SelectRel) Convert() Rel {
	return SelectRel{
		FieldName: sc.FieldName,
		Junctive:  sc.Junctive,
		T:         ConvertTypePtr(sc.T),
	}
}
func SelectSyntax(fieldName string, junctive Junctive) syntax.Word {
	return syntax.Word{
		Strings:  []string{fieldName, ""},
		Specials: []syntax.WordSpecial{syntax.WordSpecialJunct(junctive)},
	}
}
func (sc SelectRel) Syntax() Syntax {
	return SyntaxWord(SelectSyntax(sc.FieldName, sc.Junctive))
}
func (sc SelectRel) String() string { return sc.Syntax().String() }
