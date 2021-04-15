package coms

import . "don/core"

func I(theType DType) Com { return ICom(theType) }

type ICom DType

func (ic ICom) InputType() DType  { return DType(ic) }
func (ic ICom) OutputType() DType { return DType(ic) }

func (ic ICom) MeetTypes(inputType, outputType DType) Com {
	t := DType(ic)
	t.Meets(inputType)
	t.Meets(outputType)
	if t.LTE(NullType) {
		return Null
	} else {
		return ICom(t)
	}
}

func (ic ICom) Underdefined() Error {
	return DType(ic).Underdefined().Context("in I")
}

func (ic ICom) Copy() Com { return ic }

func (ic ICom) Invert() Com { return ic }

func SetEq(tcb TypedComBuilder /* mutated */, inputMap, outputMap TypeMap) {
	if inputMap.Unit != nil {
		tcb.Equate(inputMap.Unit, outputMap.Unit)
	}
	for fieldName, subInputMap := range inputMap.Fields {
		SetEq(tcb, subInputMap, outputMap.Fields[fieldName])
	}
}
func (ICom) TypedCom(tcb TypedComBuilder /* mutated */, inputMap, outputMap TypeMap) {
	SetEq(tcb, inputMap, outputMap)
}
