package com

import (
	. "don/junctive"
	"don/syntax"
)

func Select(junctive Junctive, fieldName string) Com {
	t := PairTypePtr()
	UnifyTypePtrs(TypePtrAt(junctive, fieldName, GetLeft(t)), GetRight(t))
	return SelectCom{
		FieldName: fieldName,
		Junctive:  junctive,
		T:         t,
	}
}

type SelectCom struct {
	FieldName string
	Junctive
	T *TypePtr
}

func (sc SelectCom) Type() *TypePtr { return sc.T }
func (sc SelectCom) Copy(mapping map[*TypePtr]*TypePtr) Com {
	sc.T = CopyTypePtr(sc.T, mapping)
	return sc
}
func (sc SelectCom) Convert() Com {
	return SelectCom{
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
func (sc SelectCom) Syntax() Syntax {
	return SyntaxWord(SelectSyntax(sc.FieldName, sc.Junctive))
}
func (sc SelectCom) String() string { return sc.Syntax().String() }
