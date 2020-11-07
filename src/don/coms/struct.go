package coms

import . "don/core"

type StructCom struct{}

func (StructCom) Types(inputType, outputType *DType) (underdefined Error) {
	inputType.Meets(StructType)
	inputType.Meets(*outputType)
	*outputType = *inputType
	return inputType.Underdefined().Context("in input to struct")
}

func (StructCom) Run(inputType, outputType DType, input Input, output Output) {
	ICom{}.Run(inputType, outputType, input, output)
}
