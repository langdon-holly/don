package coms

import . "don/core"

type UnitCom struct{}

func (UnitCom) Types(inputType, outputType *DType) (underdefined Error) {
	inputType.Meets(UnitType)
	inputType.Meets(*outputType)
	*outputType = *inputType
	return
}

func (UnitCom) Run(inputType, outputType DType, input Input, output Output) {
	ICom{}.Run(inputType, outputType, input, output)
}
