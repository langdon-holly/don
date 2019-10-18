package core

// DType

type DTypeTag int

const (
	UnitTypeTag = DTypeTag(iota)
	SyntaxTypeTag
	GenComTypeTag
	StructTypeTag
)

type DType struct {
	Tag   DTypeTag
	Extra interface{}
}

// Extras

type Unit struct{}

type SyntaxTag int

const (
	StringSyntaxTag = SyntaxTag(iota)
	LolSyntaxTag
	MCallSyntaxTag
)

type String string

type Lol [][]Syntax

type MCall struct {
	Macro Syntax /* String or MCall */
	Arg   Syntax /* String or Lol */
}

type Syntax struct {
	Tag   SyntaxTag
	Extra interface{}
}

type Struct map[string]interface{}

// Get DType

var UnitType = DType{UnitTypeTag, nil}

var SyntaxType = DType{SyntaxTypeTag, nil}

var GenComType = DType{GenComTypeTag, nil}

func MakeStructType(fields map[string]DType) DType {
	return DType{StructTypeTag, fields}
}
