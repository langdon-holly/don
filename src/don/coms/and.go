package coms

import . "don/core"

type AndCom struct{}

var andComInputTypeFields map[string]DType = make(map[string]DType, 2)
var andComInputType DType = MakeStructType(andComInputTypeFields)

func init() {
	andComInputTypeFields["a"] = BoolType
	andComInputTypeFields["b"] = BoolType
}

func (com AndCom) InputType() DType {
	return andComInputType
}

func (com AndCom) OutputType() DType {
	return BoolType
}

func (com AndCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	i := input.(Struct)
	a := i["a"].(Struct)
	b := i["b"].(Struct)
	aTrue := a["true"].(<-chan Unit)
	aFalse := a["false"].(<-chan Unit)
	bTrue := b["true"].(<-chan Unit)
	bFalse := b["false"].(<-chan Unit)

	o := output.(Struct)
	oTrue := o["true"].(chan<- Unit)
	oFalse := o["false"].(chan<- Unit)

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

func GenAnd(inputType DType) Com { /* TODO: Check inputType */ return AndCom{} }
