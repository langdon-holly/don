package coms

import . "don/core"

type WhenRefCom struct{}

func (WhenRefCom) OutputType(inputType DType) DType {
	switch inputType.Lvl {
	case UnknownTypeLvl:
	case NormalTypeLvl:
		if inputType.Tag != RefTypeTag {
			return ImpossibleType
		}
	case ImpossibleTypeLvl:
		return ImpossibleType
	}

	return UnitType
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
