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

func (com ChooseCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	i := input.(Struct)
	iA := i["a"].(<-chan Unit)
	iB := i["b"].(<-chan Unit)
	ready := i["ready"].(<-chan Unit)

	o := output.(Struct)
	oA := o["a"].(chan<- Unit)
	oB := o["b"].(chan<- Unit)

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

func GenChoose(inputType DType) Com { /* TODO: Check inputType */ return ChooseCom{} }
