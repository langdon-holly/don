package types

import . "don/core"

func MakeMaybeType(valType DType) DType {
	maybeType := MakeNFieldsType(2)
	maybeType.Fields["not?"] = UnitType
	maybeType.Fields["val"] = valType
	return maybeType
}
