package coms

import . "don/core"

func RunI(theType DType, inputGetter InputGetter, outputGetter OutputGetter) {
	switch theType.Tag {
	case UnitTypeTag:
		inputGetter.Unit <- <-outputGetter.Unit
	case RefTypeTag:
		inputGetter.Ref <- <-outputGetter.Ref
	case StructTypeTag:
		for fieldName, fieldType := range theType.Fields {
			go RunI(fieldType, inputGetter.Struct[fieldName], outputGetter.Struct[fieldName])
		}
	}
	return
}

type ICom struct{}

func (ICom) OutputType(inputType DType) DType { return inputType }

func (ICom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	RunI(inputType, inputGetter, outputGetter)
}
