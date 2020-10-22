package types

import . "don/core"

var BoolType DType = MakeNStructType(2)

func WriteBool(output Output, val bool) {
	if val {
		output.Fields["T"].WriteUnit()
	} else {
		output.Fields["F"].WriteUnit()
	}
}

func ReadBool(input Input) bool {
	select {
	case <-input.Fields["T"].Unit:
		return true
	case <-input.Fields["F"].Unit:
		return false
	}
}

func init() {
	BoolType.Fields["T"] = UnitType
	BoolType.Fields["F"] = UnitType
}
