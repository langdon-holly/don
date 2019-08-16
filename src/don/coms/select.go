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

func (com SelectCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	i := input.(Struct)
	for fieldName, fieldType := range com.Fields {
		if fieldName == com.FieldName {
			go RunI(fieldType, i[fieldName], output, quit)
		} else {
			go RunSink(fieldType, i[fieldName], quit)
		}
	}
}

func GenSelect(fieldName string) GenCom {
	return func(inputType DType) Com {
		if inputType.Tag != StructTypeTag {
			panic("Type error")
		}
		return SelectCom{inputType.Extra.(map[string]DType), fieldName}
	}
}
