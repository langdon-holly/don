package rel

import (
	. "don/junctive"
	"don/syntax"
)

func Collect(junctive Junctive, fieldName string) Rel {
	v := PairVarPtr()
	UnifyVarPtrs(VarGetLeft(v), VarPtrAt(junctive, fieldName, VarGetRight(v)))
	return CollectRel{
		FieldName: fieldName,
		Junctive:  junctive,
		V:         v,
	}
}

type CollectRel struct {
	FieldName string
	Junctive
	V *VarPtr
}

func (cc CollectRel) Type() *TypePtr { return VarPtrTypePtr(cc.V) }
func (cc CollectRel) Copy(mapping map[*TypePtr]*TypePtr) Rel {
	cc.V = varPtrTo(CopyTypePtr(VarPtrTypePtr(cc.V), mapping))
	return cc
}
func (cc CollectRel) Convert() Rel {
	return CollectRel{
		FieldName: cc.FieldName,
		Junctive:  cc.Junctive,
		V:         ConvertVarPtr(cc.V),
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
