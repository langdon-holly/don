package coms

import (
	. "don/core"
	"don/types"
)

func InverseYet() Com { return InverseYetCom{Yet: Yet()} }

type InverseYetCom struct{ Yet Com }

func (iyc InverseYetCom) InputType() DType { return iyc.Yet.OutputType() }

func (iyc InverseYetCom) OutputType() DType { return iyc.Yet.InputType() }

func (iyc InverseYetCom) MeetTypes(inputType, outputType DType) Com {
	iyc.Yet = iyc.Yet.MeetTypes(outputType, inputType)
	if _, nullp := iyc.Yet.(NullCom); nullp {
		return Null
	} else {
		return iyc
	}
}

func (iyc InverseYetCom) Underdefined() Error {
	return iyc.Yet.Underdefined().Context("in inverse yet")
}

func (iyc InverseYetCom) Copy() Com { iyc.Yet = iyc.Yet.Copy(); return iyc }

func (iyc InverseYetCom) Invert() Com { return iyc.Yet }

func (iyc InverseYetCom) Run(input Input, output Output) {
	inputType := iyc.Yet.OutputType()
	outputType := iyc.Yet.InputType()
	if !types.BoolType.LTE(inputType) || !yetComInputType.LTE(outputType) {
		return
	}

	com := Pipe([]Com{
		Scatter(),
		Par([]Com{
			Pipe([]Com{Select("T"), Fork(), I(yetComInputType)}),
			Pipe([]Com{Select("F"), Deselect("?")})},
		),
		Merge(),
	}).MeetTypes(inputType, outputType)
	if underdefined := com.Underdefined(); underdefined != nil {
		panic("Unreachable underdefined:\n" + underdefined.String())
	} else if !inputType.LTE(com.InputType()) {
		panic("Unreachable")
	} else if !outputType.LTE(com.OutputType()) {
		panic("Unreachable")
	}
	com.Run(input, output)
}
