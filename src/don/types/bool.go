package types

import . "don/core"

var BoolType DType = MakeNFieldsType(2)

func init() {
	BoolType.Fields["T"] = UnitType
	BoolType.Fields["F"] = UnitType
}

func WriteBool(output Output, val bool) {
	if val {
		output.Fields["T"].Converge()
	} else {
		output.Fields["F"].Converge()
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
