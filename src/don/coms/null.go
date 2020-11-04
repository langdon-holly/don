package coms

import . "don/core"

type NullCom struct{}

func (NullCom) Types(inputType, outputType *DType) (done bool) {
	inputType.Meets(NullType)
	outputType.Meets(NullType)
	return true
}

func (NullCom) Run(inputType, outputType DType, input Input, output Output) {}
