package coms

import . "don/core"

type ConstRefCom struct {
	ReferentType DType
	Val          Ref
}

func (crc ConstRefCom) OutputType(inputType PartialType) PartialType {
	return PartializeType(MakeRefType(crc.ReferentType))
}

func (crc ConstRefCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	for {
		select {
		case <-input.Unit:
			output.WriteRef(crc.Val)
		case <-quit:
			return
		}
	}
}
