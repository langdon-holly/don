package coms

import . "don/core"

type ChooseCom struct{}

var chooseComInputTypeFields map[string]DType = make(map[string]DType, 3)
var chooseComInputType DType = MakeStructType(chooseComInputTypeFields)

func init() {
	chooseComInputTypeFields["a"] = UnitType
	chooseComInputTypeFields["b"] = UnitType
	chooseComInputTypeFields["ready"] = UnitType
}

func (com ChooseCom) InputType() DType {
	return chooseComInputType
}

var chooseComOutputTypeFields map[string]DType = make(map[string]DType, 2)
var chooseComOutputType DType = MakeStructType(chooseComOutputTypeFields)

func init() {
	chooseComOutputTypeFields["a"] = UnitType
	chooseComOutputTypeFields["b"] = UnitType
}

func (com ChooseCom) OutputType() DType {
	return chooseComOutputType
}

func (com ChooseCom) Run(input Input, output Output, quit <-chan struct{}) {
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

type GenChoose struct{}

func (GenChoose) OutputType(inputType PartialType) PartialType {
	return PartializeType(chooseComOutputType)
}

func (GenChoose) Com(inputType DType) Com { return ChooseCom{} }
