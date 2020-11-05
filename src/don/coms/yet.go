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

func (YetCom) Types(inputType, outputType *DType) (done bool) {
	inputType.Meets(yetComInputType)
	outputType.Meets(types.BoolType)
	if !yetComInputType.LTE(*inputType) || !types.BoolType.LTE(*outputType) {
		*inputType = NullType
		*outputType = NullType
	}
	return true
}

func (YetCom) Run(inputType, outputType DType, input Input, output Output) {
	if !yetComInputType.LTE(inputType) || !types.BoolType.LTE(outputType) {
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
