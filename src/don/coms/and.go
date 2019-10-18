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

func (com AndCom) Run(input Input, output Output, quit <-chan struct{}) {
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

type GenAnd struct{}

func (GenAnd) OutputType(inputType PartialType) PartialType { return PartializeType(BoolType) }
func (GenAnd) Com(inputType DType) Com                      { return AndCom{} }
