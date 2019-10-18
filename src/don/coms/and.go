package coms

import . "don/core"

type AndCom struct{}

func (AndCom) OutputType(inputType PartialType) PartialType { return PartializeType(BoolType) }

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
	oTrue := o["true"].Unit
	oFalse := o["false"].Unit

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
			oTrue <- Unit{}
		} else {
			oFalse <- Unit{}
		}
	}
}
