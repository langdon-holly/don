package rel

import (
	. "don/junctive"
	"don/syntax"
)

// *VarPtr

type VarPtr interface{}
type varPtrRoot struct {
	tp *TypePtr

	// Subsets of TypePtrType(tp)'s juncts
	conjuncts map[string]*VarPtr
	disjuncts map[string]*VarPtr
}
type varPtrChild struct{ parent *VarPtr }

func varPtrGetRoot(r *VarPtr /* mutated */) (*VarPtr /* varPtrRoot */, varPtrRoot) {
	switch rAlt := (*r).(type) {
	case varPtrRoot:
		return r, rAlt
	case varPtrChild:
		vp, root := varPtrGetRoot(rAlt.parent)
		*r = varPtrChild{vp}
		return vp, root
	default:
		panic(*r)
	}
}
func varPtrTo(tp *TypePtr) *VarPtr {
	vp := VarPtr(varPtrRoot{
		tp:        tp,
		conjuncts: make(map[string]*VarPtr, 0),
		disjuncts: make(map[string]*VarPtr, 0),
	})
	return &vp
}
func unifyVarPtrJuncts(rJuncts, sJuncts map[string]*VarPtr /* mutated */) map[string]*VarPtr {
	for fieldName, fieldVarPtrS := range sJuncts {
		if fieldVarPtrR, ok := rJuncts[fieldName]; ok {
			fieldRVP, fieldRRoot := varPtrGetRoot(fieldVarPtrR)
			fieldSVP, fieldSRoot := varPtrGetRoot(fieldVarPtrS)
			if fieldRVP != fieldSVP {
				*fieldRVP, *fieldSVP = unifyVarPtrRootsWithoutType(fieldRRoot, fieldSRoot)
			}
		}
		rJuncts[fieldName] = fieldVarPtrS
	}
	return rJuncts
}
func unifyVarPtrRootsWithoutType(rRoot, sRoot varPtrRoot /* mutated */) (r, s VarPtr) {
	rRoot.conjuncts = unifyVarPtrJuncts(rRoot.conjuncts, sRoot.conjuncts)
	rRoot.disjuncts = unifyVarPtrJuncts(rRoot.disjuncts, sRoot.disjuncts)
	union := VarPtr(rRoot)
	return varPtrChild{&union}, varPtrChild{&union}
}
func UnifyVarPtrs(r, s *VarPtr /* mutated */) {
	rVP, rRoot := varPtrGetRoot(r)
	sVP, sRoot := varPtrGetRoot(s)
	if rVP != sVP {
		UnifyTypePtrs(rRoot.tp, sRoot.tp)
		*rVP, *sVP = unifyVarPtrRootsWithoutType(rRoot, sRoot)
	}
}
func VarPtrTypePtr(r *VarPtr /* mutated */) *TypePtr {
	_, rRoot := varPtrGetRoot(r)
	return rRoot.tp
}

func CopyVarPtr(r *VarPtr, varMap map[*VarPtr]*VarPtr /* mutated */, typeMap map[*TypePtr]*TypePtr /* mutated */) *VarPtr {
	vp, root := varPtrGetRoot(r)
	if _, inMapping := varMap[vp]; !inMapping {
		varMap[vp] = varPtrTo(CopyTypePtr(root.tp, typeMap))
	}
	return varMap[vp]
}

// *TypePtr

type TypePtr interface{}
type typePtrRoot struct{ t Type }
type typePtrChild struct{ parent *TypePtr }

func typePtrGetRoot(r *TypePtr /* mutated */) (*TypePtr /* typePtrRoot */, typePtrRoot) {
	switch rAlt := (*r).(type) {
	case typePtrRoot:
		return r, rAlt
	case typePtrChild:
		tp, root := typePtrGetRoot(rAlt.parent)
		*r = typePtrChild{tp}
		return tp, root
	default:
		panic(*r)
	}
}
func typePtrTo(t Type) *TypePtr {
	tp := TypePtr(typePtrRoot{t: t})
	return &tp
}
func UnifyTypePtrs(r, s *TypePtr /* mutated */) {
	rTP, rRoot := typePtrGetRoot(r)
	sTP, sRoot := typePtrGetRoot(s)
	if rTP != sTP {
		unionTP := typePtrTo(joinTypes(rRoot.t, sRoot.t))
		*rTP = typePtrChild{unionTP}
		*sTP = typePtrChild{unionTP}
	}
}
func TypePtrType(r *TypePtr /* mutated */) Type {
	_, rRoot := typePtrGetRoot(r)
	return rRoot.t
}

func CopyTypePtr(r *TypePtr, mapping map[*TypePtr]*TypePtr /* mutated */) *TypePtr {
	tp, root := typePtrGetRoot(r)
	if _, inMapping := mapping[tp]; !inMapping {
		mapping[tp] = typePtrTo(root.t.copy(mapping))
	}
	return mapping[tp]
}

// Type

type Type struct {
	Unit      bool
	Conjuncts map[string]*TypePtr
	Disjuncts map[string]*TypePtr
}

// Get Var

func AnyVarPtr() *VarPtr { return varPtrTo(AnyTypePtr()) }

func VarPtrAt(junctive Junctive, fieldName string, vp *VarPtr) *VarPtr {
	return varPtrTo(TypePtrAt(junctive, fieldName, VarPtrTypePtr(vp)))
}
func VarAtLeft(t *VarPtr) *VarPtr  { return VarPtrAt(ConJunctive, "0", t) }
func VarAtRight(t *VarPtr) *VarPtr { return VarPtrAt(ConJunctive, "1", t) }

func VarGet(fieldName string, junctive Junctive, vp *VarPtr) *VarPtr {
	_, root := varPtrGetRoot(vp)

	var juncts map[string]*VarPtr
	if junctive == ConJunctive {
		juncts = root.conjuncts
	} else /* junctive == DisJunctive */ {
		juncts = root.disjuncts
	}

	if fieldVarPtr, ok := juncts[fieldName]; ok {
		return fieldVarPtr
	} else {
		fieldVarPtr = varPtrTo(Get(fieldName, junctive, root.tp))
		juncts[fieldName] = fieldVarPtr
		return fieldVarPtr
	}
}
func VarGetLeft(v *VarPtr) *VarPtr  { return VarGet("0", ConJunctive, v) }
func VarGetRight(v *VarPtr) *VarPtr { return VarGet("1", ConJunctive, v) }

func PairVarPtr() *VarPtr {
	v := VarAtLeft(AnyVarPtr())
	UnifyVarPtrs(v, VarAtRight(AnyVarPtr()))
	return v
}

// May not round-trip
func ConvertVarPtr(v *VarPtr) *VarPtr {
	w := VarAtLeft(VarGetRight(v))
	UnifyVarPtrs(w, VarAtRight(VarGetLeft(v)))
	return w
}

// Get Type

func AnyTypePtr() *TypePtr {
	return typePtrTo(
		Type{Conjuncts: make(map[string]*TypePtr, 0), Disjuncts: make(map[string]*TypePtr, 0)},
	)
}

func TypePtrAt(junctive Junctive, fieldName string, tp *TypePtr) *TypePtr {
	juncts := make(map[string]*TypePtr, 1)
	juncts[fieldName] = tp
	var conjuncts, disjuncts map[string]*TypePtr
	if junctive == ConJunctive {
		conjuncts = juncts
		disjuncts = make(map[string]*TypePtr, 0)
	} else /* junctive == DisJunctive */ {
		conjuncts = make(map[string]*TypePtr, 0)
		disjuncts = juncts
	}
	return typePtrTo(Type{Conjuncts: conjuncts, Disjuncts: disjuncts})
}
func AtLeft(t *TypePtr) *TypePtr  { return TypePtrAt(ConJunctive, "0", t) }
func AtRight(t *TypePtr) *TypePtr { return TypePtrAt(ConJunctive, "1", t) }

func Get(fieldName string, junctive Junctive, tp *TypePtr) *TypePtr {
	return TypePtrType(tp).get(junctive, fieldName)
}
func GetLeft(t *TypePtr) *TypePtr  { return Get("0", ConJunctive, t) }
func GetRight(t *TypePtr) *TypePtr { return Get("1", ConJunctive, t) }

func PairTypePtr() *TypePtr {
	t := AtLeft(AnyTypePtr())
	UnifyTypePtrs(t, AtRight(AnyTypePtr()))
	return t
}

func (t Type) get(junctive Junctive, fieldName string) *TypePtr {
	var juncts map[string]*TypePtr
	if junctive == ConJunctive {
		juncts = t.Conjuncts
	} else /* junctive == DisJunctive */ {
		juncts = t.Disjuncts
	}
	if fieldTypePtr, ok := juncts[fieldName]; ok {
		return fieldTypePtr
	} else {
		panic("Nonexistent field")
	}
}

// May not round-trip
func ConvertTypePtr(t *TypePtr) *TypePtr {
	u := AtLeft(GetRight(t))
	UnifyTypePtrs(u, AtRight(GetLeft(t)))
	return u
}

// Other

func joinTypeJuncts(juncts0, juncts1 map[string]*TypePtr /* mutated */) map[string]*TypePtr {
	for fieldName, fieldTypePtr1 := range juncts1 {
		if fieldTypePtr0, ok := juncts0[fieldName]; ok {
			UnifyTypePtrs(fieldTypePtr0, fieldTypePtr1)
		}
		juncts0[fieldName] = fieldTypePtr1
	}
	return juncts0
}
func joinTypes(t0, t1 Type /* consumed in places where both are junctive */) Type {
	t0.Unit = t0.Unit || t1.Unit
	t0.Conjuncts = joinTypeJuncts(t0.Conjuncts, t1.Conjuncts)
	t0.Disjuncts = joinTypeJuncts(t0.Disjuncts, t1.Disjuncts)
	return t0
}

func (t Type) copy(mapping map[*TypePtr]*TypePtr) Type {
	conjuncts := make(map[string]*TypePtr, len(t.Conjuncts))
	for fieldName, fieldTypePtr := range t.Conjuncts {
		conjuncts[fieldName] = CopyTypePtr(fieldTypePtr, mapping)
	}
	disjuncts := make(map[string]*TypePtr, len(t.Disjuncts))
	for fieldName, fieldTypePtr := range t.Disjuncts {
		disjuncts[fieldName] = CopyTypePtr(fieldTypePtr, mapping)
	}
	return Type{Unit: t.Unit, Conjuncts: conjuncts, Disjuncts: disjuncts}
}

// Mutates
func (t Type) Syntax() Syntax {
	var typeConjuncts []Syntax
	if t.Unit {
		typeConjuncts = append(typeConjuncts, NameSyntax("unit"))
	}
	if len(t.Conjuncts) > 0 {
		var conjunctWordses [][]syntax.Word
		for fieldName, fieldTypePtr := range t.Conjuncts {
			conjunctWordses = append(
				conjunctWordses,
				[]syntax.Word{
					CollectSyntax(fieldName, ConJunctive),
					TypePtrType(fieldTypePtr).Syntax().Word(),
					SelectSyntax(fieldName, ConJunctive),
				},
			)
		}
		typeConjuncts = append(typeConjuncts, JunctionSyntax(ConJunctive, conjunctWordses))
	}
	if len(t.Disjuncts) > 0 {
		var disjunctWordses [][]syntax.Word
		for fieldName, fieldTypePtr := range t.Disjuncts {
			disjunctWordses = append(
				disjunctWordses,
				[]syntax.Word{
					CollectSyntax(fieldName, DisJunctive),
					TypePtrType(fieldTypePtr).Syntax().Word(),
					SelectSyntax(fieldName, DisJunctive),
				},
			)
		}
		typeConjuncts = append(typeConjuncts, JunctionSyntax(DisJunctive, disjunctWordses))
	}
	if len(typeConjuncts) == 0 {
		return NameSyntax("I")
	} else if len(typeConjuncts) == 1 {
		return typeConjuncts[0]
	} else {
		typeConjunctWordses := make([][]syntax.Word, len(typeConjuncts))
		for i, typeConjunct := range typeConjuncts {
			typeConjunctWordses[i] = typeConjunct.Composition()
		}
		return JunctionSyntax(ConJunctive, typeConjunctWordses)
	}
}

// Mutates
func (t Type) String() string { return t.Syntax().String() }
