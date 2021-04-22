package coms

import (
	. "don/core"
	"don/syntax"
)

func Deselect(fieldName string) Com {
	outputType := MakeNFieldsType(1)
	outputType.Fields[fieldName] = UnknownType
	return DeselectCom{FieldName: fieldName, outputType: outputType}
}

type DeselectCom struct {
	FieldName  string
	outputType DType
}

func (dc DeselectCom) InputType() DType {
	return dc.outputType.Get(dc.FieldName)
}
func (dc DeselectCom) OutputType() DType { return dc.outputType }
func (dc DeselectCom) MeetTypes(inputType, outputType DType) Com {
	dc.outputType.Meets(inputType.AtHigh(dc.FieldName))
	dc.outputType.Meets(outputType)
	if dc.outputType.LTE(NullType) {
		return Null
	} else {
		return dc
	}
}
func (dc DeselectCom) Underdefined() Error {
	return dc.outputType.Underdefined().Context(
		"in output from deselect field " + syntax.EscapeFieldName(dc.FieldName))
}
func (dc DeselectCom) Copy() Com { return dc }
func (dc DeselectCom) Invert() Com {
	return SelectCom{
		FieldName: dc.FieldName,
		inputType: dc.outputType,
	}
}
func (dc DeselectCom) TypedCom(tcb TypedComBuilder /* mutated */, inputMap, outputMap TypeMap) {
	SetEq(tcb, inputMap, outputMap.Fields[dc.FieldName])
}
