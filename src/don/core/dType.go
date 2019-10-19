package core

// DType

type DTypeTag int

const (
	UnitTypeTag = DTypeTag(iota)
	RefTypeTag
	StructTypeTag
)

type DType struct {
	Tag      DTypeTag
	Referent *DType           /* for Tag == RefTypeTag */
	Fields   map[string]DType /* for Tag == StructTypeTag */
}

// Get DType

var UnitType = DType{Tag: UnitTypeTag}

func MakeRefType(referentType DType) DType {
	return DType{Tag: RefTypeTag, Referent: &referentType}
}

func MakeStructType(fields map[string]DType) DType {
	return DType{Tag: StructTypeTag, Fields: fields}
}
