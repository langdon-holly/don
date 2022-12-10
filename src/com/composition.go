package com

func Composition(factors []Com) Com {
	pc := CompositionCom{
		Factors: factors,
		T:       PairTypePtr(),
	}
	for i, factor := range pc.Factors {
		UnifyTypePtrs(factor.Type(), PairTypePtr())
		if i >= 1 {
			UnifyTypePtrs(GetRight(pc.Factors[i-1].Type()), GetLeft(factor.Type()))
		}
	}
	if len(factors) == 0 {
		UnifyTypePtrs(GetLeft(pc.T), GetRight(pc.T))
	} else {
		UnifyTypePtrs(GetLeft(pc.T), GetLeft(pc.Factors[0].Type()))
		UnifyTypePtrs(GetRight(pc.T), GetRight(pc.Factors[len(pc.Factors)-1].Type()))
	}
	return pc
}

type CompositionCom struct {
	Factors []Com
	T       *TypePtr
}

func (pc CompositionCom) Type() *TypePtr { return pc.T }
func (pc CompositionCom) Copy(mapping map[*TypePtr]*TypePtr) Com {
	factors := make([]Com, len(pc.Factors))
	for i, factor := range pc.Factors {
		factors[i] = factor.Copy(mapping)
	}
	return CompositionCom{Factors: factors, T: CopyTypePtr(pc.T, mapping)}
}
func (pc CompositionCom) Convert() Com {
	factorConverses := make([]Com, len(pc.Factors))
	for i, factor := range pc.Factors {
		factorConverses[len(pc.Factors)-1-i] = factor.Convert()
	}
	return CompositionCom{
		Factors: factorConverses,
		T:       ConvertTypePtr(pc.T),
	}
}
func (pc CompositionCom) Syntax() Syntax {
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
func (pc CompositionCom) String() string { return pc.Syntax().String() }
