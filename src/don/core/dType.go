package core

import (
	"strconv"
	"strings"
)

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
var StructType = DType{NoUnit: true}
var NullType = DType{NoUnit: true, Positive: true}

func NullPtr() *DType { nt := NullType; return &nt }

func MakeNStructType(nFields int) DType {
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

func typeString(out *strings.Builder, t DType, indent []byte) {
	subIndent := append(indent, byte('\t'))
	out.WriteString("(\n")
	if !t.NoUnit {
		out.Write(subIndent)
		out.WriteString("unit\n")
	}
	if t.Positive {
		for fieldName, fieldType := range t.Fields {
			out.Write(subIndent)
			out.WriteString(fieldName)
			out.WriteString(":!")
			typeString(out, fieldType, subIndent)
			out.WriteString("\n")
		}
	} else {
		subSubIndent := append(subIndent, byte('\t'))
		out.WriteString("negative!(\n")
		for fieldName, fieldType := range t.Fields {
			out.Write(subSubIndent)
			out.WriteString(fieldName)
			out.WriteString(":!")
			typeString(out, fieldType, subIndent)
			out.WriteString("\n")
		}
		out.Write(subIndent)
		out.WriteString(")\n")
	}
	out.Write(indent)
	out.WriteString(")")
}
func (t DType) String() string {
	var b strings.Builder
	typeString(&b, t, nil)
	return b.String()
}

func FanAffineTypes(many, one *DType) Error {
	if one.LTE(NullType) {
		*many = NullType
	} else {
		many.RemakeFields()

		if many.Meets(StructType); many.Positive {
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
	return many.Underdefined()
}

// Mutates many
func FanLinearTypes(many []DType, one *DType) (underdefined Error) {
	join := NullType
	for i := range many {
		many[i].Meets(*one)
		meet := join
		meet.Meets(many[i])
		underdefined.Ors(
			meet.Nonnull().Context(
				"in meet with " + strconv.Itoa(i) + "'th of many (double use)")).Ors(
			many[i].Underdefined().Context("in " + strconv.Itoa(i) + "'th of many"))
		join.Joins(many[i])
	}
	*one = join
	return
}
