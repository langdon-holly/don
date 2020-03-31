package coms

import . "don/core"

type ConstRefCom struct {
	ReferentType DType
	Val          Ref
}

func (crc ConstRefCom) OutputType(inputType DType) (outputType DType, impossible bool) {
	if inputType.Tag != UnknownTypeTag && inputType.Tag != UnitTypeTag {
		impossible = true
	} else {
		outputType = MakeRefType(crc.ReferentType)
	}
	return
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
