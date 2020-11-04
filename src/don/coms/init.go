package coms

import . "don/core"

type InitCom struct{}

// Violates multiplicative annihilation!!
func (InitCom) Types(inputType, outputType *DType) (done bool) {
	inputType.Meets(NullType)
	outputType.Meets(UnitType)
	return true
}

func (InitCom) Run(inputType, outputType DType, input Input, output Output) {
	if !outputType.NoUnit {
		output.WriteUnit()
	}
}
