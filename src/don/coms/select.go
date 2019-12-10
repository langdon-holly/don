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

func (gc SelectCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	for fieldName, fieldType := range inputType.Fields {
		if fieldName == string(gc) {
			go RunI(fieldType, inputGetter.Struct[fieldName], outputGetter)
		} else {
			Sink.Run(fieldType, inputGetter.Struct[fieldName], OutputGetter{Struct: make(map[string]OutputGetter, 0)}, quit)
		}
	}
}
