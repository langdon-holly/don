package coms

import . "don/core"

type ConstRefCom struct {
	ReferentType DType
	Val          Ref
}

func (crc ConstRefCom) OutputType(inputType DType) DType {
	return MakeRefType(crc.ReferentType)
}

func (crc ConstRefCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	input := inputGetter.GetInput(UnitType)
	output := outputGetter.GetOutput(MakeRefType(crc.ReferentType))

	for {
		select {
		case <-input.Unit:
			output.WriteRef(crc.Val)
		case <-quit:
			return
		}
	}
}
