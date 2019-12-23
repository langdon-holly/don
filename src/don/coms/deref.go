package coms

import . "don/core"

type DerefCom struct{}

func (DerefCom) OutputType(inputType DType) DType {
	if inputType.Lvl != NormalTypeLvl {
		return inputType
	}
	if inputType.Tag != RefTypeTag {
		return ImpossibleType
	}
	return *inputType.Referent
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
