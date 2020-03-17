package coms

import . "don/core"

type SelectCom string

func (sc SelectCom) OutputType(inputType DType) DType {
	if inputType.Lvl != NormalTypeLvl {
		return inputType
	}
	if inputType.Tag != StructTypeTag {
		return ImpossibleType
	}
	return inputType.Fields[string(sc)]
}

func (sc SelectCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	for fieldName, fieldType := range inputType.Fields {
		if fieldName == string(sc) {
			go RunI(fieldType, inputGetter.Struct[fieldName], outputGetter, quit)
		} else {
			go Sink.Run(fieldType, inputGetter.Struct[fieldName], OutputGetter{Struct: make(map[string]OutputGetter, 0)}, quit)
		}
	}
}
