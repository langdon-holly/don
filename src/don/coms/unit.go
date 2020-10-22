package coms

import . "don/core"

type UnitCom struct{}

func (UnitCom) Types(inputType, outputType *DType) (bad []string, done bool) {
	done = true
	*inputType, bad = MergeTypes(*inputType, UnitType)
	if bad != nil {
		bad = append(bad, "in nonunit unit input")
		return
	}
	*outputType, bad = MergeTypes(*outputType, UnitType)
	if bad != nil {
		bad = append(bad, "in nonunit unit output")
	}
	return
}

func (UnitCom) Run(inputType, outputType DType, input Input, output Output) {
	ICom{}.Run(inputType, outputType, input, output)
}
