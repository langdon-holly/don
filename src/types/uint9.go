package types

import . "don/core"

var Uint9Type = MakeNFieldsType(9)

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

func ReadUint9(rMap ReadMap) (val int) {
	for i, fieldName := range []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"} {
		select {
		case <-rMap.Fields[fieldName].Fields["0"].Unit:
		case <-rMap.Fields[fieldName].Fields["1"].Unit:
			val += 1 << i
		}
	}
	return
}

func WriteUint9(wMap WriteMap, val int) {
	for _, fieldName := range []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"} {
		if val%2 == 0 {
			wMap.Fields[fieldName].Fields["0"].Unit <- struct{}{}
		} else {
			wMap.Fields[fieldName].Fields["1"].Unit <- struct{}{}
		}
		val /= 2
	}
}
