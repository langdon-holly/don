package coms

import . "don/core"

type DerefCom struct{}

func (DerefCom) OutputType(inputType PartialType) PartialType {
	return *inputType.Referent
}

func (DerefCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	outputType := *inputType.Referent
	input := inputGetter.GetInput(inputType)
	output := outputGetter.GetOutput(outputType)

	for {
		select {
		case val := <-input.Ref:
			if val.P {
				val.InputGetter.SendOutput(outputType, output)
			}
		case <-quit:
			return
		}
	}
}
