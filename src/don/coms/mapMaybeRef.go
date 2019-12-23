package coms

import (
	. "don/core"
	"don/types"
)

type MapMaybeRefCom struct{ Com }

func (mmrc MapMaybeRefCom) OutputType(inputType DType) DType {
	if inputType.Lvl == NormalTypeLvl {
		if inputType.Tag == StructTypeTag {
			inputType = inputType.Fields["val"]
			if inputType.Lvl == NormalTypeLvl {
				if inputType.Tag == RefTypeTag {
					inputType = *inputType.Referent
				} else {
					inputType = ImpossibleType
				}
			}
		} else {
			inputType = ImpossibleType
		}
	}
	return types.MakeMaybeType(MakeRefType(mmrc.Com.OutputType(inputType)))
}

func (mmrc MapMaybeRefCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	outputType := types.MakeMaybeType(MakeRefType(mmrc.Com.OutputType(*inputType.Fields["val"].Referent)))

	input := inputGetter.GetInput(inputType)
	notpInput := input.Struct["not?"]
	valInput := input.Struct["val"]

	output := outputGetter.GetOutput(outputType)
	notpOutput := output.Struct["not?"]
	valOutput := output.Struct["val"]

	var subquit chan struct{}

	for {
		select {
		case <-notpInput.Unit:
			if subquit != nil {
				close(subquit)
			}
			subquit = nil

			notpOutput.WriteUnit()
		case val := <-valInput.Ref:
			if subquit != nil {
				close(subquit)
			}
			subquit = make(chan struct{})

			outputIGetter, outputOGetter := MakeIO(*outputType.Referent)
			go mmrc.Run(*inputType.Referent, val.InputGetter, outputOGetter, subquit)

			valOutput.WriteRef(Ref{InputGetter: outputIGetter})
		case <-quit:
			return
		}
	}
}
