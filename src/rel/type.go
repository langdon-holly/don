package rel

import (
	. "don/junctive"
	"don/syntax"
)

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
		t := joinTypes(rRoot.t, sRoot.t)
		tp := TypePtr(typePtrRoot{t: t})
		unionTP := &tp
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

func joinTypes(t0, t1 Type /* consumed in places where both are junctive */) Type {
	t0.Unit = t0.Unit || t1.Unit
	for fieldName, fieldTypePtr1 := range t1.Conjuncts {
		if fieldTypePtr0, ok := t0.Conjuncts[fieldName]; ok {
			UnifyTypePtrs(fieldTypePtr0, fieldTypePtr1)
		}
		t0.Conjuncts[fieldName] = fieldTypePtr1
	}
	for fieldName, fieldTypePtr1 := range t1.Disjuncts {
		if fieldTypePtr0, ok := t0.Disjuncts[fieldName]; ok {
			UnifyTypePtrs(fieldTypePtr0, fieldTypePtr1)
		}
		t0.Disjuncts[fieldName] = fieldTypePtr1
	}
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
