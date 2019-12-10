package coms

import . "don/core"

type RefIsNotCom struct{}

func (RefIsNotCom) OutputType(inputType PartialType) PartialType {
	return PartializeType(UnitType)
}

func (RefIsNotCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	input := inputGetter.GetInput(inputType)
	output := outputGetter.GetOutput(UnitType)

	for {
		select {
		case val := <-input.Ref:
			if !val.P {
				output.WriteUnit()
			}
		case <-quit:
			return
		}
	}
}
