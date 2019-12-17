package coms

import (
	. "don/core"
	"don/types"
)

type MapMaybeRefCom struct{ Com }

func (mrc MapMaybeRefCom) OutputType(inputType DType) DType {
	var innerInputType DType
	if inputType.P && inputType.Fields["val"].P {
		innerInputType = *inputType.Fields["val"].Referent
	}
	return types.MakeMaybeType(MakeRefType(mrc.OutputType(innerInputType)))
}

func (mrc MapMaybeRefCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	outputType := types.MakeMaybeType(MakeRefType(mrc.OutputType(*inputType.Fields["val"].Referent)))

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
			go mrc.Run(*inputType.Referent, val.InputGetter, outputOGetter, subquit)

			valOutput.WriteRef(Ref{InputGetter: outputIGetter})
		case <-quit:
			return
		}
	}
}
