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
	if bad = inputType.Meets(YetComInputType); bad != nil {
		bad = append(bad, "in yet input type")
	} else if bad = outputType.Meets(types.BoolType); bad != nil {
		bad = append(bad, "in yet output type")
	}
	done = true
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
