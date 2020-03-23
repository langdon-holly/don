package types

import . "don/core"

var BitTypeFields map[string]DType = make(map[string]DType, 2)
var BitType DType = MakeStructType(BitTypeFields)

func init() {
	BitTypeFields["0"] = UnitType
	BitTypeFields["1"] = UnitType
}
