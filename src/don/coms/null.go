package coms

import . "don/core"

type NullCom struct{}

func (NullCom) Types(inputType, outputType *DType) (underdefined Error) {
	inputType.Meets(NullType)
	outputType.Meets(NullType)
	return
}

func (NullCom) Run(inputType, outputType DType, input Input, output Output) {}
