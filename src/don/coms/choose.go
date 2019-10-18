package coms

import . "don/core"

var chooseComOutputTypeFields map[string]DType = make(map[string]DType, 2)
var chooseComOutputType DType = MakeStructType(chooseComOutputTypeFields)

func init() {
	chooseComOutputTypeFields["a"] = UnitType
	chooseComOutputTypeFields["b"] = UnitType
}

type ChooseCom struct{}

func (ChooseCom) OutputType(inputType PartialType) PartialType {
	return PartializeType(chooseComOutputType)
}

func (ChooseCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	i := input.Struct
	iA := i["a"].Unit
	iB := i["b"].Unit
	ready := i["ready"].Unit

	o := output.Struct
	oA := o["a"].Unit
	oB := o["b"].Unit

	for {
		select {
		case <-ready:
		case <-quit:
			return
		}
		select {
		case <-iA:
			oA <- Unit{}
		case <-iB:
			oB <- Unit{}
		case <-quit:
			return
		}
	}
}
