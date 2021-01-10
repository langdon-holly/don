package coms

import (
	. "don/core"
	"don/syntax"
)

func Select(fieldName string) Com {
	inputType := MakeNFieldsType(1)
	inputType.Fields[fieldName] = UnknownType
	return &SelectCom{FieldName: fieldName, inputType: inputType}
}

type SelectCom struct {
	FieldName             string
	inputType, outputType DType
}

func (sc *SelectCom) InputType() *DType  { return &sc.inputType }
func (sc *SelectCom) OutputType() *DType { return &sc.outputType }
func (sc *SelectCom) Types() Com {
	sc.outputType.Meets(sc.inputType.Get(sc.FieldName))
	sc.inputType.Meets(sc.outputType.At(sc.FieldName))
	if sc.outputType.LTE(NullType) {
		return Null
	} else {
		return sc
	}
}
func (sc SelectCom) Underdefined() Error {
	return sc.outputType.Underdefined().Context(
		"in output from select field " + syntax.EscapeFieldName(sc.FieldName))
}
func (sc SelectCom) Copy() Com { return &sc }
func (sc SelectCom) Invert() Com {
	return &DeselectCom{
		FieldName:  sc.FieldName,
		inputType:  sc.outputType,
		outputType: sc.inputType,
	}
}
func (sc SelectCom) Run(input Input, output Output) {
	if len(sc.inputType.Fields) > 0 {
		RunI(sc.outputType, input.Fields[sc.FieldName], output)
	}
}
