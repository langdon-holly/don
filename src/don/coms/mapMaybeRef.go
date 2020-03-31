package coms

import (
	. "don/core"
	"don/types"
)

type MapMaybeRefCom struct{ Com }

func (mmrc MapMaybeRefCom) OutputType(inputType DType) (outputType DType, impossible bool) {
	if inputType.Tag != UnknownTypeTag {
		if inputType.Tag != StructTypeTag {
			impossible = true
			return
		}

		inputType = inputType.Fields["val"]
		if inputType.Tag != UnknownTypeTag {
			if inputType.Tag != RefTypeTag {
				impossible = true
				return
			}

			inputType = *inputType.Referent
		}
	}
	var subOutputType DType
	subOutputType, impossible = mmrc.Com.OutputType(inputType)
	if !impossible {
		outputType = types.MakeMaybeType(MakeRefType(subOutputType))
	}
	return
}

func (mmrc MapMaybeRefCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	subOutputType, _ := mmrc.Com.OutputType(*inputType.Fields["val"].Referent)
	outputType := types.MakeMaybeType(MakeRefType(subOutputType))

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
