package types

import . "don/core"

var BoolType DType = MakeNFieldsType(2)

func init() {
	BoolType.Fields["T"] = UnitType
	BoolType.Fields["F"] = UnitType
}

func WriteBool(wMap WriteMap, val bool) {
	if val {
		wMap.Fields["T"].Unit <- struct{}{}
	} else {
		wMap.Fields["F"].Unit <- struct{}{}
	}
}

func ReadBool(rMap ReadMap) bool {
	select {
	case <-rMap.Fields["T"].Unit:
		return true
	case <-rMap.Fields["F"].Unit:
		return false
	}
}
