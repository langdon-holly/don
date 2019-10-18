package coms

import . "don/core"

type SelectCom string

func (gc SelectCom) OutputType(inputType PartialType) PartialType {
	if inputType.P {
		return inputType.Fields[string(gc)]
	} else {
		return PartialType{}
	}
}

func (gc SelectCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	for fieldName, fieldType := range inputType.Fields {
		if fieldName == string(gc) {
			go RunI(fieldType, input.Struct[fieldName], output, quit)
		} else {
			go RunSink(fieldType, input.Struct[fieldName], quit)
		}
	}
}
