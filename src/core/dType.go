package core

import . "don/syntax"

// DType

// DTypes form a Boolean algebra
// Conceptually, a DType is a set of []string
type DType struct {
	NoUnit   bool
	Positive bool
	Fields   map[string]DType
}

// Get DType

var UnknownType DType
var UnitType = DType{Positive: true}
var FieldsType = DType{NoUnit: true}
var NullType = DType{NoUnit: true, Positive: true}

func NullPtr() *DType { nt := NullType; return &nt }

func MakeNFieldsType(nFields int) DType {
	return DType{NoUnit: true, Positive: true, Fields: make(map[string]DType, nFields)}
}

func (t DType) Get(fieldName string) DType {
	if fieldType, ok := t.Fields[fieldName]; ok {
		return fieldType
	} else if t.Positive {
		return NullType
	} else {
		return UnknownType
	}
}

func (t DType) againstPath(pathType DType) DType {
	fields := make(map[string]DType, len(t.Fields))
	for fieldName, fieldType := range t.Fields {
		fields[fieldName] = fieldType.againstPath(pathType)
	}
	u := DType{NoUnit: true, Positive: t.Positive, Fields: fields}
	if !t.NoUnit {
		u.Joins(pathType)
	}
	return u
}

// If there are negative fields, returns upper bound
func (t DType) AgainstPath(path []string) DType {
	pathType := UnitType
	for i := len(path) - 1; i >= 0; i-- {
		pathType = pathType.AtLow(path[i])
	}
	return t.againstPath(pathType)
}

// Other

func (t *DType) RemakeFields() {
	fields := make(map[string]DType, len(t.Fields))
	for fieldName, fieldType := range t.Fields {
		fields[fieldName] = fieldType
	}
	t.Fields = fields
}

// Less Than or Equal
func (t0 DType) LTE(t1 DType) bool {
	if !t0.NoUnit && t1.NoUnit || !t0.Positive && t1.Positive {
		return false
	}
	for fieldName, fieldType0 := range t0.Fields {
		if !fieldType0.LTE(t1.Get(fieldName)) {
			return false
		}
	}
	if !t0.Positive {
		for fieldName, fieldType1 := range t1.Fields {
			if _, ok := t0.Fields[fieldName]; !ok && !UnknownType.LTE(fieldType1) {
				return false
			}
		}
	}
	return true
}

func (t0 DType) Equal(t1 DType) bool { return t0.LTE(t1) && t1.LTE(t0) }

func (t DType) Complement() (c DType) {
	c.NoUnit = !t.NoUnit
	c.Positive = !t.Positive
	c.Fields = make(map[string]DType, len(t.Fields))
	for fieldName, fieldType := range t.Fields {
		c.Fields[fieldName] = fieldType.Complement()
	}
	return
}

func (t0 *DType) Meets(t1 DType) {
	t0.NoUnit = t0.NoUnit || t1.NoUnit
	if t0.Positive {
		t0.RemakeFields()
		for fieldName, fieldType0 := range t0.Fields {
			if fieldType0.Meets(t1.Get(fieldName)); fieldType0.LTE(NullType) {
				delete(t0.Fields, fieldName)
			} else {
				t0.Fields[fieldName] = fieldType0
			}
		}
	} else if t1.Positive {
		fields := make(map[string]DType)
		for fieldName, fieldType1 := range t1.Fields {
			if fieldType1.Meets(t0.Get(fieldName)); !fieldType1.LTE(NullType) {
				fields[fieldName] = fieldType1
			}
		}
		t0.Positive = true
		t0.Fields = fields
	} else {
		t0.RemakeFields()
		for fieldName, fieldType1 := range t1.Fields {
			fieldType1.Meets(t0.Get(fieldName))
			t0.Fields[fieldName] = fieldType1
		}
	}
	return
}

func (t DType) AtHigh(fieldName string) DType {
	fields := make(map[string]DType, 1)
	fields[fieldName] = t
	return DType{Fields: fields}
}
func (t DType) AtLow(fieldName string) DType {
	u := MakeNFieldsType(1)
	u.Fields[fieldName] = t
	return u
}

func (t0 *DType) Joins(t1 DType) {
	t0C := t0.Complement()
	t0C.Meets(t1.Complement())
	*t0 = t0C.Complement()
}

func (t DType) Underdefined() Error {
	if !t.Positive {
		return NewError("Negative fields")
	}
	for fieldName, fieldType := range t.Fields {
		if subUnderdefined := fieldType.Underdefined(); subUnderdefined != nil {
			return subUnderdefined.InField(fieldName)
		}
	}
	return nil
}

func (t DType) Syntax() Syntax {
	if !t.NoUnit && !t.Positive && len(t.Fields) == 0 {
		return ISyntax{}
	}

	var lFactors []Syntax
	if !t.NoUnit {
		lFactors = append(lFactors, Named{Name: "unit"})
	}
	if !t.Positive {
		cFactors := []Syntax{Named{Name: "fields"}}
		for fieldName := range t.Fields {
			cFactors = append(cFactors, Application{
				Com: Named{Name: "withoutField"},
				Arg: Quote{Named{Name: fieldName}},
			})
		}
		if len(cFactors) > 1 {
			lFactors =
				append(lFactors, Composition{cFactors})
		} else if lFactors = append(lFactors, cFactors[0]); true {
		}
	}
	for fieldName, fieldType := range t.Fields {
		lFactors =
			append(lFactors, Composition{[]Syntax{
				Named{RightMarker: true, Name: fieldName},
				fieldType.Syntax(),
				Named{LeftMarker: true, Name: fieldName},
			}})
	}
	if len(lFactors) == 1 {
		return lFactors[0]
	} else {
		return Conjunction{lFactors}
	}
}

func (t DType) String() string { return t.Syntax().String() }
