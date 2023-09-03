package rel

import (
	. "don/junctive"
	"don/syntax"
)

func Junction(junctive Junctive, juncts []Rel) Rel {
	jc := JunctionRel{
		Junctive: junctive,
		Juncts:   juncts,
		V:        AnyVarPtr(),
	}
	for _, junct := range juncts {
		UnifyTypePtrs(VarPtrTypePtr(jc.V), VarPtrTypePtr(junct.Var()))
	}
	return jc
}

type JunctionRel struct {
	Junctive
	Juncts []Rel
	V      *VarPtr
}

func (jc JunctionRel) Var() *VarPtr { return jc.V }

func (jc JunctionRel) Copy(varMap map[*VarPtr]*VarPtr, typeMap map[*TypePtr]*TypePtr) Rel {
	juncts := make([]Rel, len(jc.Juncts))
	for i, junct := range jc.Juncts {
		juncts[i] = junct.Copy(varMap, typeMap)
	}
	jc.Juncts = juncts
	jc.V = CopyVarPtr(jc.V, varMap, typeMap)
	return jc
}
func (jc JunctionRel) Convert() Rel {
	for i, junct := range jc.Juncts {
		jc.Juncts[i] = junct.Convert()
	}
	jc.V = ConvertVarPtr(jc.V)
	return jc
}
func JunctionSyntax(junctive Junctive, juncts [][]syntax.Word /* each []Word non-nil */) Syntax {
	if len(juncts) == 0 {
		if junctive == ConJunctive {
			return NameSyntax("true")
		} else {
			return NameSyntax("false")
		}
	} else {
		var opName string
		if junctive == ConJunctive {
			opName = ","
		} else {
			opName = ";"
		}
		return SyntaxWords(syntax.Words{
			Compositions: append([][]syntax.Word{NameSyntax(opName).Composition()}, juncts...),
		})
	}
}
func (jc JunctionRel) Syntax() Syntax {
	junctWordses := make([][]syntax.Word, len(jc.Juncts)) /* each []Word non-nil */
	for i, junct := range jc.Juncts {
		junctWordses[i] = junct.Syntax().Composition()
	}
	return JunctionSyntax(jc.Junctive, junctWordses)
}
func (jc JunctionRel) String() string { return jc.Syntax().String() }
