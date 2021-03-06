package types

import . "don/core"

var Uint8Type = MakeNFieldsType(8)

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

func ReadUint8(rMap ReadMap) (val int) {
	for i, fieldName := range []string{"0", "1", "2", "3", "4", "5", "6", "7"} {
		select {
		case <-rMap.Fields[fieldName].Fields["0"].Unit:
		case <-rMap.Fields[fieldName].Fields["1"].Unit:
			val += 1 << i
		}
	}
	return
}

func WriteUint8At(wMap WriteMap, val int, path []string) {
	for _, fieldName := range []string{"0", "1", "2", "3", "4", "5", "6", "7"} {
		var digit WriteMap
		if val%2 == 0 {
			digit = wMap.Fields[fieldName].Fields["0"]
		} else {
			digit = wMap.Fields[fieldName].Fields["1"]
		}
		for _, fieldName := range path {
			digit = digit.Fields[fieldName]
		}
		digit.Unit <- struct{}{}
		val /= 2
	}
}
func WriteUint8(wMap WriteMap, val int) { WriteUint8At(wMap, val, nil) }
