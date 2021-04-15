package types

import . "don/core"

var BitType DType = MakeNFieldsType(2)

func init() {
	BitType.Fields["0"] = UnitType
	BitType.Fields["1"] = UnitType
}

func WriteBit(wMap WriteMap, val int) {
	if val == 0 {
		wMap.Fields["0"].Unit <- struct{}{}
	} else {
		wMap.Fields["1"].Unit <- struct{}{}
	}
}
