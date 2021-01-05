package coms

import (
	. "don/core"
	"don/types"
)

type YetCom struct{}

var yetComInputType = MakeNFieldsType(2)

func init() {
	yetComInputType.Fields[""] = UnitType
	yetComInputType.Fields["?"] = UnitType
}

func (YetCom) Instantiate() ComInstance {
	return &yetInstance{
		inputType:  yetComInputType,
		outputType: types.BoolType}
}

func (YetCom) Inverse() Com { return InverseYetCom{} }

type yetInstance struct{ inputType, outputType DType }

func (yi *yetInstance) InputType() *DType  { return &yi.inputType }
func (yi *yetInstance) OutputType() *DType { return &yi.outputType }

func (yi *yetInstance) Types() {
	if !yetComInputType.LTE(yi.inputType) || !types.BoolType.LTE(yi.outputType) {
		yi.inputType = NullType
		yi.outputType = NullType
	}
}

func (yi yetInstance) Underdefined() Error { return nil }

func (yi yetInstance) Run(input Input, output Output) {
	if !yetComInputType.LTE(yi.inputType) || !types.BoolType.LTE(yi.outputType) {
		return
	}
	<-input.Fields["?"].Unit
	select {
	case <-input.Fields[""].Unit:
		output.Fields["T"].Converge()
	default:
		output.Fields["F"].Converge()
	}
}
