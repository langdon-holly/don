package core

import "strings"

// DType

type DTypeTag int

const (
	UnknownTypeTag DTypeTag = iota
	UnitTypeTag
	StructTypeTag
)

type DType struct {
	Tag DTypeTag

	// for Tag == StructTypeTag
	Fields map[string]DType
}

// Get DType

var UnknownType DType
var UnitType = DType{Tag: UnitTypeTag}

func MakeNStructType(nFields int) DType {
	return DType{Tag: StructTypeTag, Fields: make(map[string]DType, nFields)}
}

// Other

func (theType *DType) RemakeFields() {
	fields := make(map[string]DType, len(theType.Fields))
	for fieldName, fieldType := range theType.Fields {
		fields[fieldName] = fieldType
	}
	theType.Fields = fields
}

func (t0 DType) Equal(t1 DType) bool {
	if t0.Tag != t1.Tag || len(t0.Fields) != len(t1.Fields) {
		return false
	}
	for fieldName, fieldType0 := range t0.Fields {
		fieldType1, exists := t1.Fields[fieldName]
		if !exists || !fieldType0.Equal(fieldType1) {
			return false
		}
	}
	return true
}

func mergeTags(t0, t1 DTypeTag) (merged DTypeTag, bad []string) {
	if t0 == UnknownTypeTag {
		merged = t1
	} else if t1 == UnknownTypeTag {
		merged = t0
	} else if t0 != t1 {
		bad = []string{"Cannot be both unit and struct"}
	} else {
		merged = t0
	}
	return
}

func (t0 *DType) Meets(t1 DType) (bad []string) {
	var tag DTypeTag
	if tag, bad = mergeTags(t0.Tag, t1.Tag); bad != nil || tag != StructTypeTag {
		*t0 = DType{Tag: tag}
	} else if t0.Tag == UnknownTypeTag {
		*t0 = t1
	} else if t1.Tag != UnknownTypeTag {
		t0.RemakeFields()
		for fieldName, fieldType0 := range t0.Fields {
			fieldType1, inT1 := t1.Fields[fieldName]
			if !inT1 {
				bad = []string{"Different fields"}
				return
			} else if bad = fieldType0.Meets(fieldType1); bad != nil {
				bad = append(bad, "in field "+fieldName)
				return
			}
			t0.Fields[fieldName] = fieldType0
		}
		if len(t0.Fields) < len(t1.Fields) {
			bad = []string{"Different fields"}
		}
	}
	return
}

func (t DType) Minimal() bool {
	if t.Tag != StructTypeTag {
		return t.Tag == UnitTypeTag
	}
	for _, fieldType := range t.Fields {
		if !fieldType.Minimal() {
			return false
		}
	}
	return true
}

func typeString(out *strings.Builder, t DType, indent []byte) {
	switch t.Tag {
	case UnknownTypeTag:
		out.WriteString("I")
	case UnitTypeTag:
		out.WriteString("unit")
	case StructTypeTag:
		subIndent := append(indent, byte('\t'))
		out.WriteString("(\n")
		for fieldName, fieldType := range t.Fields {
			out.Write(subIndent)
			out.WriteString(":")
			out.WriteString(fieldName)
			out.WriteString(":!")
			typeString(out, fieldType, subIndent)
			out.WriteString("\n")
		}
		out.Write(indent)
		out.WriteString(")")
	}
}
func (t DType) String() string {
	var b strings.Builder
	typeString(&b, t, nil)
	return b.String()
}

// Mutates many
func topFanTypes(many map[string]DType, one *DType) (bad []string) {
	tag := one.Tag
	for fieldName, fieldType := range many {
		if tag, bad = mergeTags(tag, fieldType.Tag); bad != nil {
			bad = append(bad, "in fanning under top-level field "+fieldName)
			return
		}
	}
	if tag == UnitTypeTag {
		for fieldName := range many {
			many[fieldName] = UnitType
		}
		*one = UnitType
	}
	return
}

func allManyFanTypes(many map[string]DType, one *DType) (bad []string) {
	if len(many) == 0 {
		bad = []string{"Empty list"}
	} else if len(many) == 1 {
		for fieldName, fieldType := range many {
			bad = one.Meets(fieldType)
			many[fieldName] = *one
		}
	} else if bad = topFanTypes(many, one); bad == nil {
		for _, fieldType := range many {
			if fieldType.Tag != StructTypeTag {
				return
			}
		}

		manyFields := make(map[string]struct{})
		for fieldName, fieldType := range many {
			for subFieldName := range fieldType.Fields {
				manyFields[subFieldName] = struct{}{}
			}

			fieldType.RemakeFields()
			many[fieldName] = fieldType
		}

		if one.Tag == StructTypeTag {
			for fieldName := range manyFields {
				if _, inOne := one.Fields[fieldName]; !inOne {
					bad = []string{"Field fans into nowhere: " + fieldName}
					return
				}
			}
			if len(manyFields) < len(one.Fields) {
				bad = []string{"Some field fans out to nowhere"}
				return
			}
			one.RemakeFields()
		} else {
			*one = MakeNStructType(len(manyFields))
			for fieldName := range manyFields {
				one.Fields[fieldName] = UnknownType
			}
		}

		// one.Tag == StructTypeTag

		for fieldName := range one.Fields {
			subMany := make(map[string]DType)
			for manyFieldName, manyFieldType := range many {
				if manyFieldFieldType, inManyField := manyFieldType.Fields[fieldName]; inManyField {
					subMany[manyFieldName] = manyFieldFieldType
				}
			}
			subOne := one.Fields[fieldName]

			if bad = allManyFanTypes(subMany, &subOne); bad != nil {
				bad = append(bad, "in fan field "+fieldName)
				return
			}

			for subManyFieldName, subManyFieldType := range subMany {
				many[subManyFieldName].Fields[fieldName] = subManyFieldType
			}
			one.Fields[fieldName] = subOne
		}
	}
	return
}

// Mutates many
func FanTypes(allMany bool, many map[string]DType, one *DType) (bad []string) {
	if allMany {
		return allManyFanTypes(many, one)
	} else {
		return topFanTypes(many, one)
	}
}

func MergeType2As(t0, t1 *DType) (bad []string) {
	bad = t0.Meets(*t1)
	*t1 = *t0
	return
}
