package coms

import . "don/core"

type InitCom struct{}

func (InitCom) Types(inputType, outputType *DType) (bad []string, done bool) {
	if bad = inputType.Meets(MakeNStructType(0)); bad != nil {
		bad = append(bad, "in bad input type for init")
	} else if bad = outputType.Meets(UnitType); bad != nil {
		bad = append(bad, "in bad output type for init")
	}
	done = true
	return
}

func (InitCom) Run(inputType, outputType DType, input Input, output Output) {
	output.WriteUnit()
}
