package core

var BoolTypeFields map[string]DType = make(map[string]DType, 2)
var BoolType DType = MakeStructType(BoolTypeFields)

func WriteBool(output Output, val bool) {
	if val {
		output.Struct["true"].Unit <- Unit{}
	} else {
		output.Struct["false"].Unit <- Unit{}
	}
}

func ReadBool(input Input) bool {
	select {
	case <-input.Struct["true"].Unit:
		return true
	case <-input.Struct["false"].Unit:
		return false
	}
}

func init() {
	BoolTypeFields["true"] = UnitType
	BoolTypeFields["false"] = UnitType
}