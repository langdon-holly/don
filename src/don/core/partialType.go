package core

type PartialType struct {
	P      bool
	Tag    DTypeTag
	Fields map[string]PartialType /* for Tag == StructTypeTag */
}

func (pt0 PartialType) Equal(pt1 PartialType) bool {
	if pt0.P != pt1.P {
		return false
	}
	if !pt0.P {
		return true
	}

	if pt0.Tag != pt1.Tag {
		return false
	}
	if pt0.Tag != StructTypeTag {
		return true
	}

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

	return true
}

func MergePartialTypes(t0, t1 PartialType) PartialType {
	if !t0.P {
		return t1
	}
	if !t1.P {
		return t0
	}
	if t0.Tag != t1.Tag {
		panic("Partial-type mismatch in merge")
	}
	if t0.Tag != StructTypeTag {
		return t0
	}
	if len(t0.Fields) != len(t1.Fields) {
		panic("Struct partial types have different len(fields) in merge")
	}

	ret := PartialType{P: true, Tag: StructTypeTag, Fields: make(map[string]PartialType, len(t0.Fields))}

	for fieldName, t0FieldPType := range t0.Fields {
		t1FieldPType, exists := t1.Fields[fieldName]
		if !exists {
			panic("Field missing from partial type in merge")
		}
		ret.Fields[fieldName] = MergePartialTypes(t0FieldPType, t1FieldPType)
	}
	return ret
}

func PartializeType(theType DType) PartialType {
	if theType.Tag == StructTypeTag {
		pFields := make(map[string]PartialType, len(theType.Fields))
		for fieldName, fieldType := range theType.Fields {
			pFields[fieldName] = PartializeType(fieldType)
		}
		return PartialType{P: true, Tag: StructTypeTag, Fields: pFields}
	} else {
		return PartialType{P: true, Tag: theType.Tag}
	}
}

func HolizePartialType(pType PartialType) DType {
	if !pType.P {
		panic("Strictly partial type in holization")
	}

	if pType.Tag == StructTypeTag {
		fields := make(map[string]DType, len(pType.Fields))
		for fieldName, fieldPType := range pType.Fields {
			fields[fieldName] = HolizePartialType(fieldPType)
		}
		return MakeStructType(fields)
	} else {
		return DType{Tag: pType.Tag}
	}
}

func PartialTypeAtPath(pType PartialType, fieldPath []string) PartialType {
	for i := len(fieldPath) - 1; i >= 0; i-- {
		fields := make(map[string]PartialType, 1)
		fields[fieldPath[i]] = pType

		pType = PartialType{P: true, Tag: StructTypeTag, Fields: fields}
	}
	return pType
}
