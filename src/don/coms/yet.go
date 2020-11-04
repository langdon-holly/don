package coms

import (
	. "don/core"
	"don/types"
)

type YetCom struct{}

var YetComInputType = MakeNStructType(2)

func init() {
	YetComInputType.Fields[""] = UnitType
	YetComInputType.Fields["?"] = UnitType
}

func (YetCom) Types(inputType, outputType *DType) (done bool) {
	if inputType.LTE(NullType) {
		*outputType = NullType
	} else if outputType.LTE(NullType) {
		*inputType = NullType
	} else if inputType.Meets(YetComInputType); true {
		outputType.Meets(types.BoolType)
	}
	return true
}

func (YetCom) Run(inputType, outputType DType, input Input, output Output) {
	itInput := input.Fields[""]
	askInput := input.Fields["?"]
	trueOutput := output.Fields["T"]
	falseOutput := output.Fields["F"]
	if itInput.Unit == nil ||
		askInput.Unit == nil ||
		trueOutput.Unit == nil ||
		falseOutput.Unit == nil {
		return
	}
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
