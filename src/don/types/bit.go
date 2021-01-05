package types

import . "don/core"

var BitType DType = MakeNFieldsType(2)

func init() {
	BitType.Fields["0"] = UnitType
	BitType.Fields["1"] = UnitType
}

func WriteBit(output Output, val int) {
	if val == 0 {
		output.Fields["0"].Converge()
	} else {
		output.Fields["1"].Converge()
	}
}
