package core

import "strings"

// DType

type DType struct {
	NoUnit   bool
	Positive bool
	Fields   map[string]DType /* for Positive */
}

// Get DType

var UnknownType DType
var UnitType = DType{Positive: true}
var StructType = DType{NoUnit: true}
var NullType = DType{NoUnit: true, Positive: true}

func MakeNStructType(nFields int) DType {
	return DType{NoUnit: true, Positive: true, Fields: make(map[string]DType, nFields)}
}

func (t DType) Get(fieldName string) DType {
	if !t.Positive {
		return UnknownType
	} else if fieldType, ok := t.Fields[fieldName]; ok {
		return fieldType
	} else {
		return NullType
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
	} else if !t1.Positive {
		return true
	}
	for fieldName, fieldType0 := range t0.Fields {
		if !fieldType0.LTE(t1.Get(fieldName)) {
			return false
		}
	}
	return true
}

func (t0 DType) Equal(t1 DType) bool { return t0.LTE(t1) && t1.LTE(t0) }

func (t0 *DType) Meets(t1 DType) {
	t0.NoUnit = t0.NoUnit || t1.NoUnit
	if !t0.Positive {
		t0.Positive = t1.Positive
		t0.Fields = t1.Fields
	} else if t1.Positive {
		t0.RemakeFields()
		for fieldName, fieldType0 := range t0.Fields {
			if fieldType0.Meets(t1.Get(fieldName)); fieldType0.LTE(NullType) {
				delete(t0.Fields, fieldName)
			} else {
				t0.Fields[fieldName] = fieldType0
			}
		}
	}
	return
}

func (t0 *DType) Joins(t1 DType) {
	t0.NoUnit = t0.NoUnit && t1.NoUnit
	if t0.Positive = t0.Positive && t1.Positive; t0.Positive {
		t0.RemakeFields()
		for fieldName, fieldType := range t1.Fields {
			fieldType.Joins(t0.Get(fieldName))
			t0.Fields[fieldName] = fieldType
		}
	} else if t0.Fields = nil; true {
	}
	return
}

func (t DType) Done() bool {
	if !t.Positive {
		return false
	}
	for _, fieldType := range t.Fields {
		if !fieldType.Done() {
			return false
		}
	}
	return true
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
		out.Write(subIndent)
		out.WriteString("struct\n")
	}
	out.Write(indent)
	out.WriteString(")")
}
func (t DType) String() string {
	var b strings.Builder
	typeString(&b, t, nil)
	return b.String()
}

// Mutates many
func FanTypes(many, one *DType) (done bool) {
	if one.LTE(NullType) {
		*many = NullType
	} else if many.Meets(StructType); many.Positive {
		many.RemakeFields()

		join := NullType
		for fieldName, fieldType := range many.Fields {
			fieldType.Meets(*one)
			many.Fields[fieldName] = fieldType
			join.Joins(fieldType)
		}
		*one = join
	}
	return many.Done()
}
