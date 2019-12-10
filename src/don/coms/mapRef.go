package coms

import . "don/core"

type MapRefCom struct {
	Com
}

func (mrc MapRefCom) OutputType(inputType PartialType) PartialType {
	return MakeRefPartialType(mrc.Com.OutputType(*inputType.Referent))
}

func (mrc MapRefCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	outputType := MakeRefType(HolizePartialType(mrc.Com.OutputType(PartializeType(*inputType.Referent))))
	input := inputGetter.GetInput(inputType)
	output := outputGetter.GetOutput(outputType)
	var subquit chan struct{}

	for {
		select {
		case val := <-input.Ref:
			if subquit != nil {
				close(subquit)
			}
			if val.P {
				subquit = make(chan struct{})

				outputIGetter, outputOGetter := MakeIO(*outputType.Referent)
				go mrc.Com.Run(*inputType.Referent, val.InputGetter, outputOGetter, subquit)

				output.WriteRef(Ref{P: true, InputGetter: outputIGetter})
			} else {
				subquit = nil

				output.WriteRef(Ref{})
			}
		case <-quit:
			return
		}
	}
}
