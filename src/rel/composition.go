package rel

func Composition(factors []Rel) Rel {
	pc := CompositionRel{
		Factors: factors,
		V:       PairVarPtr(),
	}
	for i, factor := range pc.Factors {
		UnifyVarPtrs(varPtrTo(factor.Type()), PairVarPtr())
		if i >= 1 {
			UnifyVarPtrs(VarGetRight(varPtrTo(pc.Factors[i-1].Type())), VarGetLeft(varPtrTo(factor.Type())))
		}
	}
	if len(factors) == 0 {
		UnifyVarPtrs(VarGetLeft(pc.V), VarGetRight(pc.V))
	} else {
		UnifyVarPtrs(VarGetLeft(pc.V), VarGetLeft(varPtrTo(pc.Factors[0].Type())))
		UnifyVarPtrs(VarGetRight(pc.V), VarGetRight(varPtrTo(pc.Factors[len(pc.Factors)-1].Type())))
	}
	return pc
}

type CompositionRel struct {
	Factors []Rel
	V       *VarPtr
}

func (pc CompositionRel) Type() *TypePtr { return VarPtrTypePtr(pc.V) }
func (pc CompositionRel) Copy(mapping map[*TypePtr]*TypePtr) Rel {
	factors := make([]Rel, len(pc.Factors))
	for i, factor := range pc.Factors {
		factors[i] = factor.Copy(mapping)
	}
	return CompositionRel{Factors: factors, V: varPtrTo(CopyTypePtr(VarPtrTypePtr(pc.V), mapping))}
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
		composition := make(SyntaxComposition, len(pc.Factors))
		for i, factor := range pc.Factors {
			composition[i] = factor.Syntax().Word()
		}
		return composition
	}
}
func (pc CompositionRel) String() string { return pc.Syntax().String() }
