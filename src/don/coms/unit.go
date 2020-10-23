package coms

import . "don/core"

type UnitCom struct{}

func (UnitCom) Types(inputType, outputType *DType) (bad []string, done bool) {
	if bad = inputType.Meets(UnitType); bad != nil {
		bad = append(bad, "in nonunit unit input")
	} else if bad = outputType.Meets(UnitType); bad != nil {
		bad = append(bad, "in nonunit unit output")
	}
	done = true
	return
}

func (UnitCom) Run(inputType, outputType DType, input Input, output Output) {
	ICom{}.Run(inputType, outputType, input, output)
}
