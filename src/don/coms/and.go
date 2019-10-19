package coms

import . "don/core"
import "don/types"

type AndCom struct{}

func (AndCom) OutputType(inputType PartialType) PartialType { return PartializeType(types.BoolType) }

func (AndCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	n := len(inputType.Fields)

	trues := make([]<-chan Unit, n)
	falses := make([]<-chan Unit, n)

	i := 0
	for fieldName := range inputType.Fields {
		trues[i] = input.Struct[fieldName].Struct["true"].Unit
		falses[i] = input.Struct[fieldName].Struct["false"].Unit
		i++
	}

	o := output.Struct
	oTrue := o["true"]
	oFalse := o["false"]

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
