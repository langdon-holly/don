package coms

import . "don/core"

type UnitCom struct{}

func (UnitCom) Types(inputType, outputType *DType) (done bool) {
	inputType.Meets(UnitType)
	inputType.Meets(*outputType)
	*outputType = *inputType
	return true
}

func (UnitCom) Run(inputType, outputType DType, input Input, output Output) {
	ICom{}.Run(inputType, outputType, input, output)
}
