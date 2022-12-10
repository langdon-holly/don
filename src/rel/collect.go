package rel

import (
	. "don/junctive"
	"don/syntax"
)

func Collect(junctive Junctive, fieldName string) Rel {
	t := PairTypePtr()
	UnifyTypePtrs(GetLeft(t), TypePtrAt(junctive, fieldName, GetRight(t)))
	return CollectRel{
		FieldName: fieldName,
		Junctive:  junctive,
		T:         t,
	}
}

type CollectRel struct {
	FieldName string
	Junctive
	T *TypePtr
}

func (cc CollectRel) Type() *TypePtr { return cc.T }
func (cc CollectRel) Copy(mapping map[*TypePtr]*TypePtr) Rel {
	cc.T = CopyTypePtr(cc.T, mapping)
	return cc
}
func (cc CollectRel) Convert() Rel {
	return CollectRel{
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
func (cc CollectRel) Syntax() Syntax {
	return SyntaxWord(CollectSyntax(cc.FieldName, cc.Junctive))
}
func (cc CollectRel) String() string { return cc.Syntax().String() }
