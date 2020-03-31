package coms

import . "don/core"

type DerefCom struct{}

func (DerefCom) OutputType(inputType DType) (outputType DType, impossible bool) {
	if inputType.Tag == UnknownTypeTag {
		return
	}
	if inputType.Tag != RefTypeTag {
		impossible = true
		return
	}
	outputType = *inputType.Referent
	return
}

func (DerefCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	outputType := *inputType.Referent
	input := inputGetter.GetInput(inputType)
	output := outputGetter.GetOutput(outputType)

	for {
		select {
		case val := <-input.Ref:
			val.SendOutput(outputType, output)
		case <-quit:
			return
		}
	}
}
