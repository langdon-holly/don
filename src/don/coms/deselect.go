package coms

import . "don/core"

type DeselectCom struct {
	FieldName string
	FieldType DType
}

func (com DeselectCom) InputType() DType {
	return com.FieldType
}

func (com DeselectCom) OutputType() DType {
	fields := make(map[string]DType, 1)
	fields[com.FieldName] = com.FieldType
	return MakeStructType(fields)
}

func (com DeselectCom) Run(input Input, output Output, quit <-chan struct{}) {
	RunI(com.FieldType, input, output.Struct[com.FieldName], quit)
}

type GenDeselect string

func (gd GenDeselect) OutputType(inputType PartialType) (ret PartialType) {
	ret.P = true
	ret.Tag = StructTypeTag

	ret.Fields = make(map[string]PartialType, 1)
	ret.Fields[string(gd)] = inputType

	return
}

func (gd GenDeselect) Com(inputType DType) Com {
	return DeselectCom{FieldName: string(gd), FieldType: inputType}
}
