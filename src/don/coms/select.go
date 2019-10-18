package coms

import . "don/core"

type SelectCom struct {
	Fields    map[string]DType
	FieldName string
}

func (com SelectCom) InputType() DType {
	return MakeStructType(com.Fields)
}

func (com SelectCom) OutputType() DType {
	return com.Fields[com.FieldName]
}

func (com SelectCom) Run(input Input, output Output, quit <-chan struct{}) {
	i := input.Struct
	for fieldName, fieldType := range com.Fields {
		if fieldName == com.FieldName {
			go RunI(fieldType, i[fieldName], output, quit)
		} else {
			go RunSink(fieldType, i[fieldName], quit)
		}
	}
}

type GenSelect string

func (gc GenSelect) OutputType(inputType PartialType) PartialType {
	if inputType.P {
		return inputType.Fields[string(gc)]
	} else {
		return PartialType{}
	}
}

func (gc GenSelect) Com(inputType DType) Com {
	return SelectCom{
		Fields:    inputType.Fields,
		FieldName: string(gc)}
}
