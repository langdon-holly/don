package com

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
func DebugTypePtr() *TypePtr {
	tp := TypePtr(typePtrRoot{t: AnyType{}})
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

type Type interface {
	get(Junctive, string) *TypePtr
	copy(map[*TypePtr]*TypePtr) Type
	Syntax() Syntax /* Mutates */
	String() string /* Mutates */
}

type AnyType struct{}
type JunctiveType struct {
	Junctive
	Juncts map[string]*TypePtr /* Non-nil */
}
type NoType struct{}

// Get Type

var UnitType Type = AnyType{}

func AnyTypePtr() *TypePtr { return typePtrTo(AnyType{}) }
func NoTypePtr() *TypePtr  { return typePtrTo(NoType{}) }
func TypePtrAt(junctive Junctive, fieldName string, tp *TypePtr) *TypePtr {
	u := JunctiveType{Junctive: junctive, Juncts: make(map[string]*TypePtr, 1)}
	u.Juncts[fieldName] = tp
	return typePtrTo(u)
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

// Mutates
func (AnyType) get(junctive Junctive, fieldName string) *TypePtr {
	panic("Get field of any")
}
func (jt JunctiveType) get(junctive Junctive, fieldName string) *TypePtr {
	if jt.Junctive != junctive {
		panic("Wrong junctive")
	} else if fieldTypePtr, ok := jt.Juncts[fieldName]; ok {
		return fieldTypePtr
	} else {
		panic("Nonexistent field")
	}
}
func (NoType) get(junctive Junctive, fieldName string) *TypePtr {
	return NoTypePtr()
}

// May not round-trip
func ConvertTypePtr(t *TypePtr) *TypePtr {
	u := AtLeft(GetRight(t))
	UnifyTypePtrs(u, AtRight(GetLeft(t)))
	return u
}

// Other

func joinTypes(t0, t1 Type /* consumed in places where both are junctive */) Type {
	if _, t0Any := t0.(AnyType); false {
	} else if _, t1No := t1.(NoType); t0Any || t1No {
		return t1
	} else if _, t1Any := t1.(AnyType); false {
	} else if _, t0No := t0.(NoType); t1Any || t0No {
		return t0
	} else if j0 := t0.(JunctiveType); false {
	} else if j1 := t1.(JunctiveType); j0.Junctive != j1.Junctive {
	} else {
		for fieldName, fieldTypePtr1 := range j1.Juncts {
			if fieldTypePtr0, ok := j0.Juncts[fieldName]; ok {
				UnifyTypePtrs(fieldTypePtr0, fieldTypePtr1)
			}
			j0.Juncts[fieldName] = fieldTypePtr1
		}
		return j0
	}
	return NoType{}
}

func (AnyType) copy(_ map[*TypePtr]*TypePtr) Type { return AnyType{} }
func (j JunctiveType) copy(mapping map[*TypePtr]*TypePtr) Type {
	juncts := make(map[string]*TypePtr, len(j.Juncts))
	for fieldName, fieldTypePtr := range j.Juncts {
		juncts[fieldName] = CopyTypePtr(fieldTypePtr, mapping)
	}
	return JunctiveType{Junctive: j.Junctive, Juncts: juncts}
}
func (NoType) copy(_ map[*TypePtr]*TypePtr) Type { return NoType{} }

// Mutates
func (_ AnyType) Syntax() Syntax { return NameSyntax("I") }
func (j JunctiveType) Syntax() Syntax {
	var junctWordses [][]syntax.Word
	for fieldName, fieldTypePtr := range j.Juncts {
		junctWordses = append(
			junctWordses,
			[]syntax.Word{
				CollectSyntax(fieldName, j.Junctive),
				TypePtrType(fieldTypePtr).Syntax().Word(),
				SelectSyntax(fieldName, j.Junctive),
			},
		)
	}
	return JunctionSyntax(j.Junctive, junctWordses)
}
func (_ NoType) Syntax() Syntax { return NameSyntax("false") }

// Mutates
func (t AnyType) String() string      { return t.Syntax().String() }
func (t JunctiveType) String() string { return t.Syntax().String() }
func (t NoType) String() string       { return t.Syntax().String() }
