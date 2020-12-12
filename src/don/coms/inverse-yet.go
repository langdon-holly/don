package coms

import (
	. "don/core"
	"don/types"
)

type InverseYetCom struct{}

func (InverseYetCom) Instantiate() ComInstance {
	return inverseYetInstance{Yet: YetCom{}.Instantiate()}
}

func (InverseYetCom) Inverse() Com { return YetCom{} }

type inverseYetInstance struct{ Yet ComInstance }

func (iyi inverseYetInstance) InputType() *DType { return iyi.Yet.OutputType() }

func (iyi inverseYetInstance) OutputType() *DType { return iyi.Yet.InputType() }

func (iyi inverseYetInstance) Types() { iyi.Yet.Types() }

func (iyi inverseYetInstance) Underdefined() Error {
	return iyi.Yet.Underdefined().Context("in inverse yet")
}

func (iyi inverseYetInstance) Run(input Input, output Output) {
	inputType := *iyi.Yet.OutputType()
	outputType := *iyi.Yet.InputType()
	if !types.BoolType.LTE(inputType) || !yetComInputType.LTE(outputType) {
		return
	}

	comI := PipeCom([]Com{
		ScatterCom{},
		ParCom([]Com{
			PipeCom([]Com{SelectCom("T"), ForkCom{}, ICom(yetComInputType)}),
			PipeCom([]Com{SelectCom("F"), DeselectCom("?")})}),
		MergeCom{}}).Instantiate()
	comI.InputType().Meets(inputType)
	comI.OutputType().Meets(outputType)
	comI.Types()
	if underdefined := comI.Underdefined(); underdefined != nil {
		panic("Unreachable underdefined:\n" + underdefined.String())
	} else if !inputType.LTE(*comI.InputType()) {
		panic("Unreachable")
	} else if !outputType.LTE(*comI.OutputType()) {
		panic("Unreachable")
	}
	comI.Run(input, output)
}
