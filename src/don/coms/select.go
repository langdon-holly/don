package coms

import . "don/core"
import "don/extra"

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
			_, sink := extra.MakeIOChans(fieldType, 0)
			go RunI(fieldType, input.Struct[fieldName], sink, quit)
		}
	}
}
