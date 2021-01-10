package core

import . "don/syntax"

// DType

// DTypes form a Boolean algebra
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

func (t DType) At(fieldName string) DType {
	fields := make(map[string]DType, 1)
	fields[fieldName] = t
	return DType{Fields: fields}
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

func (t DType) Nonnull() Error {
	if !t.NoUnit {
		return NewError("Unit")
	} else if !t.Positive {
		return NewError("Negative fields")
	}
	for fieldName, fieldType := range t.Fields {
		if subNonnull := fieldType.Nonnull(); subNonnull != nil {
			return subNonnull.InField(fieldName)
		}
	}
	return nil
}

func (t DType) Syntax() Syntax {
	if !t.NoUnit && !t.Positive && len(t.Fields) == 0 {
		return Syntax{Tag: ISyntaxTag}
	}

	var children []Syntax
	if !t.NoUnit {
		children = append(children, Syntax{Tag: NameSyntaxTag, Name: "unit"})
	}
	if !t.Positive {
		composed := []Syntax{{Tag: NameSyntaxTag, Name: "fields"}}
		for fieldName := range t.Fields {
			composed = append(composed, Syntax{
				Tag: ApplicationSyntaxTag,
				Children: []Syntax{
					{Tag: NameSyntaxTag, Name: "withoutField"},
					{Tag: QuotationSyntaxTag, Children: []Syntax{{
						Tag:  NameSyntaxTag,
						Name: fieldName,
					}}},
				},
			})
		}
		if len(composed) > 1 {
			children =
				append(children, Syntax{Tag: CompositionSyntaxTag, Children: composed})
		} else if children = append(children, composed[0]); true {
		}
	}
	for fieldName, fieldType := range t.Fields {
		children =
			append(children, Syntax{Tag: CompositionSyntaxTag, Children: []Syntax{
				{Tag: NameSyntaxTag, RightMarker: true, Name: fieldName},
				fieldType.Syntax(),
				{Tag: NameSyntaxTag, LeftMarker: true, Name: fieldName},
			}})
	}
	if len(children) == 0 {
		return Syntax{Tag: ListSyntaxTag}
	} else if len(children) == 1 {
		return children[0]
	} else {
		return Syntax{Tag: CompositionSyntaxTag, Children: []Syntax{
			{Tag: NameSyntaxTag, Name: "<"},
			{Tag: ListSyntaxTag, Children: children},
			{Tag: NameSyntaxTag, Name: ">"},
		}}
	}
}

func (t DType) String() string { return t.Syntax().String() }

func FanAffineTypes(many, one *DType) Error {
	if one.LTE(NullType) {
		*many = NullType
	} else {
		many.RemakeFields()
		if many.Meets(FieldsType); many.Positive {
			join := NullType
			for fieldName, fieldType := range many.Fields {
				fieldType.Meets(*one)
				many.Fields[fieldName] = fieldType
				join.Joins(fieldType)
			}
			*one = join
		} else {
			for fieldName, fieldType := range many.Fields {
				fieldType.Meets(*one)
				many.Fields[fieldName] = fieldType
			}
		}
	}
	return many.Underdefined().Context("in many")
}

func FanLinearTypes(many, one *DType) (underdefined Error) {
	if underdefined = FanAffineTypes(many, one); underdefined == nil {
		joinSoFar := NullType
		fieldsSoFar := make([]string, 0, len(many.Fields))
		for fieldName, fieldType := range many.Fields {
			meet := joinSoFar
			meet.Meets(fieldType)
			if meet.Nonnull() != nil {
				for _, prevFieldName := range fieldsSoFar {
					meet := many.Fields[prevFieldName]
					meet.Meets(fieldType)
					if nonnull := meet.Nonnull(); nonnull != nil {
						underdefined = nonnull.Context(
							"in meet of fields " +
								prevFieldName +
								" and " +
								fieldName +
								" in many (double use)")
						return
					}
				}
			}
			joinSoFar.Joins(fieldType)
			fieldsSoFar = append(fieldsSoFar, fieldName)
		}
	}
	return
}
