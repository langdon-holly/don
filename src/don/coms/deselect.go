package coms

import (
	. "don/core"
	"don/syntax"
)

func Deselect(fieldName string) Com {
	outputType := MakeNFieldsType(1)
	outputType.Fields[fieldName] = UnknownType
	return &DeselectCom{FieldName: fieldName, outputType: outputType}
}

type DeselectCom struct {
	FieldName             string
	inputType, outputType DType
}

func (dc *DeselectCom) InputType() *DType  { return &dc.inputType }
func (dc *DeselectCom) OutputType() *DType { return &dc.outputType }
func (dc *DeselectCom) Types() Com {
	dc.inputType.Meets(dc.outputType.Get(dc.FieldName))
	dc.outputType.Meets(dc.inputType.At(dc.FieldName))
	if dc.inputType.LTE(NullType) {
		return Null
	} else {
		return dc
	}
}
func (dc DeselectCom) Underdefined() Error {
	return dc.inputType.Underdefined().Context(
		"in input to deselect field " + syntax.EscapeFieldName(dc.FieldName))
}
func (dc DeselectCom) Copy() Com { return &dc }
func (dc DeselectCom) Invert() Com {
	return &SelectCom{
		FieldName:  dc.FieldName,
		inputType:  dc.outputType,
		outputType: dc.inputType,
	}
}
func (dc DeselectCom) Run(input Input, output Output) {
	if len(dc.outputType.Fields) > 0 {
		RunI(dc.inputType, input, output.Fields[dc.FieldName])
	}
}
