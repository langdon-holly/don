package core

// DType

type DTypeTag int

const (
	UnknownTypeTag = DTypeTag(iota)
	UnitTypeTag
	RefTypeTag
	StructTypeTag
)

type DType struct {
	Tag      DTypeTag
	Referent *DType           /* for Tag == RefTypeTag */
	Fields   map[string]DType /* for Tag == StructTypeTag */
}

// What TODO about partial struct types?

// Get DType

var UnknownType = DType{}

var UnitType = DType{Tag: UnitTypeTag}

func MakeRefType(referentType DType) DType {
	return DType{Tag: RefTypeTag, Referent: &referentType}
}

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

func assumingEqual(t0, t1 DType, assumedEquals map[*DType]map[*DType]struct{}) bool {
	if t0.Tag != t1.Tag {
		return false
	}

	switch t0.Tag {
	case RefTypeTag:
		rights, ok := assumedEquals[t0.Referent]
		if !ok {
			rights = make(map[*DType]struct{}, 1)
			assumedEquals[t0.Referent] = rights
		}

		_, ok = rights[t1.Referent]
		if ok {
			/* assumed equal */
			return true
		} else {
			/* assume they're equal */
			rights[t1.Referent] = struct{}{}
			return assumingEqual(*t0.Referent, *t1.Referent, assumedEquals)
		}
	case StructTypeTag:
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

func (t0 DType) Equal(t1 DType) bool {
	return assumingEqual(t0, t1, make(map[*DType]map[*DType]struct{}))
}

func recursiveMerge(t0, t1 DType, referentMerges map[*DType]map[*DType]*DType) (merged DType, impossible bool) {
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
	case RefTypeTag:
		merged = DType{Tag: RefTypeTag}

		rights, ok := referentMerges[t0.Referent]
		if !ok {
			rights = make(map[*DType]*DType, 1)
			referentMerges[t0.Referent] = rights
		}

		if merge, inProgress := rights[t1.Referent]; inProgress {
			merged.Referent = merge.Referent
		} else {
			merged.Referent = new(DType)
			rights[t1.Referent] = merged.Referent
			*merged.Referent, impossible = recursiveMerge(*t0.Referent, *t1.Referent, referentMerges)
			if impossible {
				return
			}
		}

		return
	case StructTypeTag:
		merged = DType{Tag: StructTypeTag, Fields: make(map[string]DType)}

		for fieldName, t0FieldType := range t0.Fields {
			t1FieldType := t1.Fields[fieldName]
			merged.Fields[fieldName], impossible = recursiveMerge(t0FieldType, t1FieldType, referentMerges)
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

func MergeTypes(t0, t1 DType) (merged DType, impossible bool) {
	return recursiveMerge(t0, t1, make(map[*DType]map[*DType]*DType))
}
