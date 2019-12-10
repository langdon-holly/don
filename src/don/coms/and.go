package coms

import (
	. "don/core"
	"don/types"
)

type AndCom struct{}

func (AndCom) OutputType(inputType PartialType) PartialType { return PartializeType(types.BoolType) }

func (AndCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	n := len(inputType.Fields)

	trues := make([]<-chan Unit, n)
	falses := make([]<-chan Unit, n)

	i := 0
	for fieldName := range inputType.Fields {
		trues[i] = inputGetter.Struct[fieldName].Struct["true"].GetInput(UnitType).Unit
		falses[i] = inputGetter.Struct[fieldName].Struct["false"].GetInput(UnitType).Unit
		i++
	}

	oTrue := outputGetter.Struct["true"].GetOutput(UnitType)
	oFalse := outputGetter.Struct["false"].GetOutput(UnitType)

	for {
		val := true

		for i := 0; i < n; i++ {
			select {
			case <-trues[i]:
			case <-falses[i]:
				val = false
			case <-quit:
				return
			}
		}

		if val {
			oTrue.WriteUnit()
		} else {
			oFalse.WriteUnit()
		}
	}
}
