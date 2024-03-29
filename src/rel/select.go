package rel

import (
	. "don/junctive"
	"don/syntax"
)

func Select(junctive Junctive, fieldName string) Rel {
	v := PairVarPtr()
	UnifyVarPtrs(VarPtrAt(junctive, fieldName, VarGetLeft(v)), VarGetRight(v))
	return SelectRel{
		FieldName: fieldName,
		Junctive:  junctive,
		V:         v,
	}
}

type SelectRel struct {
	FieldName string
	Junctive
	V *VarPtr
}

func (sc SelectRel) Var() *VarPtr { return sc.V }
func (sc SelectRel) Copy(varMap map[*VarPtr]*VarPtr, typeMap map[*TypePtr]*TypePtr) Rel {
	sc.V = CopyVarPtr(sc.V, varMap, typeMap)
	return sc
}
func (sc SelectRel) Convert() Rel {
	return SelectRel{
		FieldName: sc.FieldName,
		Junctive:  sc.Junctive,
		V:         ConvertVarPtr(sc.V),
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
