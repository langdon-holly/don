package types

import . "don/core"

func MakeMaybeType(valType DType) DType {
	fields := make(map[string]DType, 2)
	fields["not?"] = UnitType
	fields["val"] = valType
	return MakeStructType(fields)
}
