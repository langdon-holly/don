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

func (com DeselectCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	RunI(com.FieldType, input, output.(Struct)[com.FieldName], quit)
}

func GenDeselect(fieldName string) GenCom {
	return func(inputType DType) Com { return DeselectCom{fieldName, inputType} }
}
