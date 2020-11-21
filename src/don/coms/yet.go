package coms

import (
	. "don/core"
	"don/types"
)

type YetCom struct{}

var yetComInputType = MakeNStructType(2)

func init() {
	yetComInputType.Fields[""] = UnitType
	yetComInputType.Fields["?"] = UnitType
}

func (YetCom) Instantiate() ComInstance {
	return &yetInstance{
		inputType:  yetComInputType,
		outputType: types.BoolType}
}

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
	itInput := input.Fields[""]
	askInput := input.Fields["?"]
	trueOutput := output.Fields["T"]
	falseOutput := output.Fields["F"]
	for {
		<-askInput.Unit
		select {
		case <-itInput.Unit:
			trueOutput.WriteUnit()
		default:
			falseOutput.WriteUnit()
		}
	}
}
