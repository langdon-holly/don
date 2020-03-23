package types

import . "don/core"

var Uint8Type = MakeStructType(make(map[string]DType, 8))

func init() {
	Uint8Type.Fields["0"] = BitType
	Uint8Type.Fields["1"] = BitType
	Uint8Type.Fields["2"] = BitType
	Uint8Type.Fields["3"] = BitType
	Uint8Type.Fields["4"] = BitType
	Uint8Type.Fields["5"] = BitType
	Uint8Type.Fields["6"] = BitType
	Uint8Type.Fields["7"] = BitType
}

func WriteUint8(output Output, val int) {
	for _, fieldName := range []string{"0", "1", "2", "3", "4", "5", "6", "7"} {
		if val%2 == 0 {
			output.Struct[fieldName].Struct["0"].WriteUnit()
		} else {
			output.Struct[fieldName].Struct["1"].WriteUnit()
		}
		val /= 2
	}
}
