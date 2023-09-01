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

func (cc CollectRel) Var() *VarPtr { return cc.V }
func (cc CollectRel) Copy(varMap map[*VarPtr]*VarPtr, typeMap map[*TypePtr]*TypePtr) Rel {
	cc.V = CopyVarPtr(cc.V, varMap, typeMap)
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
