package coms

import . "don/core"

type AndCom struct{}

func (AndCom) OutputType(inputType PartialType) PartialType { return PartializeType(BoolType) }

func (AndCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	i := input.Struct
	a := i["a"].Struct
	b := i["b"].Struct
	aTrue := a["true"].Unit
	aFalse := a["false"].Unit
	bTrue := b["true"].Unit
	bFalse := b["false"].Unit

	o := output.Struct
	oTrue := o["true"].Unit
	oFalse := o["false"].Unit

	var aVal, bVal bool
	for {
		select {
		case <-aTrue:
			aVal = true
		case <-aFalse:
			aVal = false
		case <-quit:
			return
		}
		select {
		case <-bTrue:
			bVal = true
		case <-bFalse:
			bVal = false
		case <-quit:
			return
		}
		if aVal && bVal {
			oTrue <- Unit{}
		} else {
			oFalse <- Unit{}
		}
	}
}
