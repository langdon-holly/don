package extra

import . "don/core"

func Run(com Com, inputType, outputType *DType) (inputW Output, outputR Input, overdefined []string, underdefined bool) {
	notUnderdefined := true
	overdefined, notUnderdefined = com.Types(inputType, outputType)
	underdefined = !notUnderdefined
	if overdefined != nil || underdefined {
		return
	}

	inputR, inputW := MakeIO(*inputType)
	outputR, outputW := MakeIO(*outputType)

	go com.Run(*inputType, *outputType, inputR, outputW)

	return
}
