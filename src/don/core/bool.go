package core

var BoolTypeFields map[string]DType = make(map[string]DType, 2)
var BoolType DType = MakeStructType(BoolTypeFields)

func WriteBool(output interface{}, val bool) {
	if val {
		output.(Struct)["true"].(chan<- Unit) <- Unit{}
	} else {
		output.(Struct)["false"].(chan<- Unit) <- Unit{}
	}
}

func ReadBool(input interface{}) bool {
	select {
	case <-input.(Struct)["true"].(<-chan Unit):
		return true
	case <-input.(Struct)["false"].(<-chan Unit):
		return false
	}
}

func init() {
	BoolTypeFields["true"] = UnitType
	BoolTypeFields["false"] = UnitType
}
