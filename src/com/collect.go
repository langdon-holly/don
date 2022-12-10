package com

import (
	. "don/junctive"
	"don/syntax"
)

func Collect(junctive Junctive, fieldName string) Com {
	t := PairTypePtr()
	UnifyTypePtrs(GetLeft(t), TypePtrAt(junctive, fieldName, GetRight(t)))
	return CollectCom{
		FieldName: fieldName,
		Junctive:  junctive,
		T:         t,
	}
}

type CollectCom struct {
	FieldName string
	Junctive
	T *TypePtr
}

func (cc CollectCom) Type() *TypePtr { return cc.T }
func (cc CollectCom) Copy(mapping map[*TypePtr]*TypePtr) Com {
	cc.T = CopyTypePtr(cc.T, mapping)
	return cc
}
func (cc CollectCom) Convert() Com {
	return CollectCom{
		FieldName: cc.FieldName,
		Junctive:  cc.Junctive,
		T:         ConvertTypePtr(cc.T),
	}
}
func CollectSyntax(fieldName string, junctive Junctive) syntax.Word {
	return syntax.Word{
		Strings:  []string{"", fieldName},
		Specials: []syntax.WordSpecial{syntax.WordSpecialJunct(junctive)},
	}
}
func (cc CollectCom) Syntax() Syntax {
	return SyntaxWord(CollectSyntax(cc.FieldName, cc.Junctive))
}
func (cc CollectCom) String() string { return cc.Syntax().String() }
