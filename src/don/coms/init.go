package coms

import . "don/core"

type InitCom struct{}

func (InitCom) Types(inputType, outputType *DType) (bad []string, done bool) {
	done = true
	*inputType, bad = MergeTypes(*inputType, MakeNStructType(0))
	if bad != nil {
		bad = append(bad, "in bad input type for init")
		return
	}
	*outputType, bad = MergeTypes(*outputType, UnitType)
	if bad != nil {
		bad = append(bad, "in bad output type for init")
	}
	return
}

func (InitCom) Run(inputType, outputType DType, input Input, output Output) {
	output.WriteUnit()
}
