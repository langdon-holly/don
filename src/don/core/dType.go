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

//func (t DType) AtPath(fieldPath []string) DType {
//	for _, fieldName := range fieldPath {
//		t = t.Fields[fieldName]
//	}
//	return t
//}

//func TypeAtPath(theType DType, fieldPath []string) DType {
//	for i := len(fieldPath) - 1; i >= 0; i-- {
//		fields := make(map[string]DType, 1)
//		fields[fieldPath[i]] = theType
//
//		theType = DType{Fields: fields}
//	}
//	return theType
//}

// Other

func (theType *DType) RemakeFields() {
	fields := make(map[string]DType, len(theType.Fields))
	for fieldName, fieldType := range theType.Fields {
		fields[fieldName] = fieldType
	}
	theType.Fields = fields
}

func (t0 DType) Equal(t1 DType) bool {
	if t0.Tag != t1.Tag {
		return false
	}
	if len(t0.Fields) != len(t1.Fields) {
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

func MergeTags(t0, t1 DTypeTag) (merged DTypeTag, bad []string) {
	if t0 == UnknownTypeTag {
		merged = t1
		return
	}
	if t1 == UnknownTypeTag {
		merged = t0
		return
	}

	if t0 != t1 {
		bad = []string{"Cannot be both unit and struct"}
		return
	}

	merged = t0
	return
}

func MergeTypes(t0, t1 DType) (merged DType, bad []string) {
	merged.Tag, bad = MergeTags(t0.Tag, t1.Tag)
	if bad != nil || merged.Tag != StructTypeTag {
		return
	}

	if t0.Tag == UnknownTypeTag {
		merged = t1
		return
	} else if t1.Tag == UnknownTypeTag {
		merged = t0
		return
	}

	merged.Fields = make(map[string]DType, len(t0.Fields))
	for fieldName, fieldType0 := range t0.Fields {
		fieldType1, inT1 := t1.Fields[fieldName]
		if !inT1 {
			bad = []string{"Different fields"}
			return
		}
		merged.Fields[fieldName], bad = MergeTypes(fieldType0, fieldType1)
		if bad != nil {
			bad = append(bad, "in field "+fieldName)
			return
		}
	}
	if len(t0.Fields) < len(t1.Fields) {
		bad = []string{"Different fields"}
		return
	}
	return
}

func (t DType) Minimal() bool {
	if t.Tag == StructTypeTag {
		for _, fieldType := range t.Fields {
			if !fieldType.Minimal() {
				return false
			}
		}
		return true
	} else {
		return t.Tag == UnitTypeTag
	}
}

func typeString(out *strings.Builder, t DType, indent []byte) {
	switch t.Tag {
	case UnknownTypeTag:
		out.WriteRune('?')
	case UnitTypeTag:
		out.WriteRune('I')
	case StructTypeTag:
		subIndent := append(indent, byte('\t'))
		out.WriteString("(\n")
		for fieldName, fieldType := range t.Fields {
			out.Write(subIndent)
			out.WriteString(fieldName)
			out.WriteString(": ")
			typeString(out, fieldType, subIndent)
			out.WriteRune('\n')
		}
		out.Write(indent)
		out.WriteRune(')')
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
		tag, bad = MergeTags(tag, fieldType.Tag)
		if bad != nil {
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
	bad = topFanTypes(many, one)
	if bad != nil {
		return
	}

	if len(many) == 0 {
		*one, bad = MergeTypes(*one, MakeNStructType(0))
	}
	if len(many) == 1 {
		for fieldName, fieldType := range many {
			*one, bad = MergeTypes(fieldType, *one)
			many[fieldName] = *one
		}
		return
	}

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

		bad = allManyFanTypes(subMany, &subOne)
		if bad != nil {
			bad = append(bad, "in fan field "+fieldName)
			return
		}

		for subManyFieldName, subManyFieldType := range subMany {
			many[subManyFieldName].Fields[fieldName] = subManyFieldType
		}
		one.Fields[fieldName] = subOne
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
	var merged DType
	merged, bad = MergeTypes(*t0, *t1)
	if bad != nil {
		return
	}
	*t0 = merged
	*t1 = merged
	return
}

func MergeTypeAs(types []*DType) (bad []string) {
	var merged DType
	for _, aType := range types {
		merged, bad = MergeTypes(merged, *aType)
		if bad != nil {
			return
		}
	}
	for i := range types {
		*types[i] = merged
	}
	return
}
