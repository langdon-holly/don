package core

// DType

type DTypeTag int

const (
	UnitTypeTag = DTypeTag(iota)
	SyntaxTypeTag
	ComTypeTag
	StructTypeTag
)

type DType struct {
	Tag    DTypeTag
	Fields map[string]DType /* for Tag == StructTypeTag */
}

// Get DType

var UnitType = DType{Tag: UnitTypeTag}

var SyntaxType = DType{Tag: SyntaxTypeTag}

var ComType = DType{Tag: ComTypeTag}

func MakeStructType(fields map[string]DType) DType {
	return DType{Tag: StructTypeTag, Fields: fields}
}
