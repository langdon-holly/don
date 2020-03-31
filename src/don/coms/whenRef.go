package coms

import . "don/core"

type WhenRefCom struct{}

func (WhenRefCom) OutputType(inputType DType) (outputType DType, impossible bool) {
	if inputType.Tag == UnknownTypeTag || inputType.Tag == RefTypeTag {
		outputType = UnitType
	} else {
		impossible = true
	}
	return
}

func (WhenRefCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	input := inputGetter.GetInput(inputType)
	output := outputGetter.GetOutput(UnitType)

	for {
		select {
		case <-input.Ref:
			output.WriteUnit()
		case <-quit:
			return
		}
	}
}
