package core

// DType

type DTypeTag int

const (
	UnknownTypeTag = DTypeTag(iota)
	UnitTypeTag
	StructTypeTag
)

type DType struct {
	Tag    DTypeTag
	Fields map[string]DType /* for Tag == StructTypeTag */
}

// Get DType

var UnknownType = DType{}

var UnitType = DType{Tag: UnitTypeTag}

func MakeStructType(fields map[string]DType) DType {
	return DType{Tag: StructTypeTag, Fields: fields}
}

func TypeAtPath(theType DType, fieldPath []string) DType {
	for i := len(fieldPath) - 1; i >= 0; i-- {
		fields := make(map[string]DType, 1)
		fields[fieldPath[i]] = theType

		theType = MakeStructType(fields)
	}
	return theType
}

// Other

func (t0 DType) Equal(t1 DType) bool {
	if t0.Tag != t1.Tag {
		return false
	}

	if t0.Tag == StructTypeTag {
		if len(t0.Fields) != len(t1.Fields) {
			return false
		}
		for fieldName, fieldType0 := range t0.Fields {
			fieldType1, exists := t1.Fields[fieldName]
			if !exists {
				return false
			}
			if !fieldType0.Equal(fieldType1) {
				return false
			}
		}
	}

	return true
}

func MergeTypes(t0, t1 DType) (merged DType, impossible bool) {
	if t0.Tag == UnknownTypeTag {
		merged = t1
		return
	}
	if t1.Tag == UnknownTypeTag {
		merged = t0
		return
	}

	if t0.Tag != t1.Tag {
		impossible = true
		return
	}

	switch t0.Tag {
	case UnitTypeTag:
		merged = t0
		return
	case StructTypeTag:
		merged = DType{Tag: StructTypeTag, Fields: make(map[string]DType)}

		for fieldName, t0FieldType := range t0.Fields {
			t1FieldType := t1.Fields[fieldName]
			merged.Fields[fieldName], impossible = MergeTypes(t0FieldType, t1FieldType)
			if impossible {
				return
			}
		}
		for fieldName, t1FieldType := range t1.Fields {
			if _, exists := t0.Fields[fieldName]; !exists {
				merged.Fields[fieldName] = t1FieldType
			}
		}

		return
	default:
		panic("Unreachable")
	}
}
