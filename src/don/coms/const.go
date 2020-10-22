package coms

import . "don/core"

func Const(outputType DType) Com {
	comsN := len(outputType.Fields)
	if outputType.Tag == UnitTypeTag {
		comsN++
	}

	coms := make([]Com, comsN)
	i := 0
	for fieldName, fieldType := range outputType.Fields {
		coms[i] = PipeCom([]Com{Const(fieldType), DeselectCom(fieldName)})
		i++
	}
	if outputType.Tag == UnitTypeTag {
		coms[i] = ICom{}
	}

	return SplitMergeCom(coms)
}
