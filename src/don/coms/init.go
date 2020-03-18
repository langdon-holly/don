package coms

import . "don/core"

type InitCom struct{}

func (InitCom) OutputType(inputType DType) DType {
	return UnitType
}

func (InitCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	go Sink.Run(inputType, inputGetter, OutputGetter{}, quit)
	outputGetter.GetOutput(UnitType).WriteUnit()
}
