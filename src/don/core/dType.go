package core

// DType

type DTypeTag int

const (
	UnitTypeTag = DTypeTag(iota)
	RefTypeTag
	StructTypeTag
)

type DType struct {
	P        bool
	Tag      DTypeTag         /* for P */
	Referent *DType           /* for P && Tag == RefTypeTag */
	Fields   map[string]DType /* for P && Tag == StructTypeTag */
}

// Get DType

var UnknownType = DType{}

var UnitType = DType{P: true, Tag: UnitTypeTag}

func MakeRefType(referentType DType) DType {
	return DType{P: true, Tag: RefTypeTag, Referent: &referentType}
}

func MakeStructType(fields map[string]DType) DType {
	return DType{P: true, Tag: StructTypeTag, Fields: fields}
}

func TypeAtPath(pType DType, fieldPath []string) DType {
	for i := len(fieldPath) - 1; i >= 0; i-- {
		fields := make(map[string]DType, 1)
		fields[fieldPath[i]] = pType

		pType = DType{P: true, Tag: StructTypeTag, Fields: fields}
	}
	return pType
}

// Other

func assumingEqual(pt0, pt1 DType, assumedEquals map[*DType]map[*DType]struct{}) bool {
	if pt0.P != pt1.P {
		return false
	}
	if !pt0.P {
		return true
	}

	if pt0.Tag != pt1.Tag {
		return false
	}

	if pt0.Tag == RefTypeTag {
		rights, ok := assumedEquals[pt0.Referent]
		if !ok {
			rights = make(map[*DType]struct{}, 1)
			assumedEquals[pt0.Referent] = rights
		}

		_, ok = rights[pt1.Referent]
		if ok {
			/* assumed equal */
			return true
		} else {
			/* assume they're equal */
			rights[pt1.Referent] = struct{}{}
			return assumingEqual(*pt0.Referent, *pt1.Referent, assumedEquals)
		}
	} else if pt0.Tag == StructTypeTag {

		if len(pt0.Fields) != len(pt1.Fields) {
			return false
		}
		for fieldName, fieldPType0 := range pt0.Fields {
			fieldPType1, exists := pt1.Fields[fieldName]
			if !exists {
				return false
			}
			if !fieldPType0.Equal(fieldPType1) {
				return false
			}
		}
	}

	return true
}

func (pt0 DType) Equal(pt1 DType) bool {
	return assumingEqual(pt0, pt1, make(map[*DType]map[*DType]struct{}))
}

func MergeTypes(t0, t1 DType) DType {
	if !t0.P {
		return t1
	}
	if !t1.P {
		return t0
	}
	if t0.Tag != t1.Tag {
		panic("Type mismatch in merge")
	}
	if t0.Tag != StructTypeTag {
		return t0
	}
	if len(t0.Fields) != len(t1.Fields) {
		panic("Struct types have different len(fields) in merge")
	}

	ret := DType{P: true, Tag: StructTypeTag, Fields: make(map[string]DType, len(t0.Fields))}

	for fieldName, t0FieldPType := range t0.Fields {
		t1FieldPType, exists := t1.Fields[fieldName]
		if !exists {
			panic("Field missing from type in merge")
		}
		ret.Fields[fieldName] = MergeTypes(t0FieldPType, t1FieldPType)
	}
	return ret
}
