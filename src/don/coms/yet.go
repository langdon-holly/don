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

func (YetCom) Types(inputType, outputType *DType) (bad []string, done bool) {
	done = true
	*inputType, bad = MergeTypes(*inputType, YetComInputType)
	if bad != nil {
		bad = append(bad, "in yet input type")
		return
	}
	*outputType, bad = MergeTypes(*outputType, types.BoolType)
	if bad != nil {
		bad = append(bad, "in yet output type")
	}
	return
}

func (YetCom) Run(inputType, outputType DType, input Input, output Output) {
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
