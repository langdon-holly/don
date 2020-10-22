package types

import . "don/core"

var Uint9Type = MakeNStructType(9)

func init() {
	Uint9Type.Fields["0"] = BitType
	Uint9Type.Fields["1"] = BitType
	Uint9Type.Fields["2"] = BitType
	Uint9Type.Fields["3"] = BitType
	Uint9Type.Fields["4"] = BitType
	Uint9Type.Fields["5"] = BitType
	Uint9Type.Fields["6"] = BitType
	Uint9Type.Fields["7"] = BitType
	Uint9Type.Fields["8"] = BitType
}

func ReadUint9(input Input) (val int) {
	for i, fieldName := range []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"} {
		select {
		case <-input.Fields[fieldName].Fields["0"].Unit:
		case <-input.Fields[fieldName].Fields["1"].Unit:
			val += 1 << i
		}
	}
	return
}

func WriteUint9(output Output, val int) {
	for _, fieldName := range []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"} {
		if val%2 == 0 {
			output.Fields[fieldName].Fields["0"].WriteUnit()
		} else {
			output.Fields[fieldName].Fields["1"].WriteUnit()
		}
		val /= 2
	}
}
