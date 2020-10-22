package types

import . "don/core"

func MakeMaybeType(valType DType) DType {
	maybeType := MakeNStructType(2)
	maybeType.Fields["not?"] = UnitType
	maybeType.Fields["val"] = valType
	return maybeType
}
