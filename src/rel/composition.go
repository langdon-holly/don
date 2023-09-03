package rel

func Composition(factors []Rel) Rel {
	pc := CompositionRel{
		Factors: factors,
		V:       PairVarPtr(),
	}
	for i, factor := range pc.Factors {
		UnifyVarPtrs(factor.Var(), PairVarPtr())
		if i >= 1 {
			UnifyVarPtrs(VarGetRight(pc.Factors[i-1].Var()), VarGetLeft(factor.Var()))
		}
	}
	if len(factors) == 0 {
		UnifyVarPtrs(VarGetLeft(pc.V), VarGetRight(pc.V))
	} else {
		UnifyVarPtrs(VarGetLeft(pc.V), VarGetLeft(pc.Factors[0].Var()))
		UnifyVarPtrs(VarGetRight(pc.V), VarGetRight(pc.Factors[len(pc.Factors)-1].Var()))
	}
	return pc
}

type CompositionRel struct {
	Factors []Rel
	V       *VarPtr
}

func (pc CompositionRel) Var() *VarPtr { return pc.V }
func (pc CompositionRel) Copy(varMap map[*VarPtr]*VarPtr, typeMap map[*TypePtr]*TypePtr) Rel {
	factors := make([]Rel, len(pc.Factors))
	for i, factor := range pc.Factors {
		factors[i] = factor.Copy(varMap, typeMap)
	}
	return CompositionRel{Factors: factors, V: CopyVarPtr(pc.V, varMap, typeMap)}
}
func (pc CompositionRel) Convert() Rel {
	factorConverses := make([]Rel, len(pc.Factors))
	for i, factor := range pc.Factors {
		factorConverses[len(pc.Factors)-1-i] = factor.Convert()
	}
	return CompositionRel{
		Factors: factorConverses,
		V:       ConvertVarPtr(pc.V),
	}
}
func (pc CompositionRel) Syntax() Syntax {
	if len(pc.Factors) == 0 {
		return NameSyntax("I")
	} else {
		composition := make(SyntaxComposition, len(pc.Factors)) /* non-nil */
		for i, factor := range pc.Factors {
			composition[i] = factor.Syntax().Word()
		}
		return composition
	}
}
func (pc CompositionRel) String() string { return pc.Syntax().String() }
