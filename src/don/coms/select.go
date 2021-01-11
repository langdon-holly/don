package coms

import (
	. "don/core"
	"don/syntax"
)

func Select(fieldName string) Com {
	inputType := MakeNFieldsType(1)
	inputType.Fields[fieldName] = UnknownType
	return SelectCom{FieldName: fieldName, inputType: inputType}
}

type SelectCom struct {
	FieldName string
	inputType DType
}

func (sc SelectCom) InputType() DType  { return sc.inputType }
func (sc SelectCom) OutputType() DType { return sc.inputType.Get(sc.FieldName) }
func (sc SelectCom) MeetTypes(inputType, outputType DType) Com {
	sc.inputType.Meets(inputType)
	sc.inputType.Meets(outputType.At(sc.FieldName))
	if sc.inputType.LTE(NullType) {
		return Null
	} else {
		return sc
	}
}
func (sc SelectCom) Underdefined() Error {
	return sc.inputType.Underdefined().Context(
		"in input to select field " + syntax.EscapeFieldName(sc.FieldName))
}
func (sc SelectCom) Copy() Com { return sc }
func (sc SelectCom) Invert() Com {
	return DeselectCom{
		FieldName:  sc.FieldName,
		outputType: sc.inputType,
	}
}
func (sc SelectCom) Run(input Input, output Output) {
	if len(sc.inputType.Fields) > 0 {
		RunI(sc.inputType.Get(sc.FieldName), input.Fields[sc.FieldName], output)
	}
}
