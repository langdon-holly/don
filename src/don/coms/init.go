package coms

import . "don/core"

type InitCom struct{}

func (InitCom) OutputType(inputType DType) DType {
	return UnitType
}

func (InitCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	inputGetter.GetInput(inputType)
	outputGetter.GetOutput(UnitType).WriteUnit()
}
