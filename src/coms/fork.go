package coms

import . "don/core"

func Fork() Com { return ForkCom{outputType: FieldsType} }

type ForkCom struct {
	inputType, outputType DType
	underdefined          Error
}

func (fc ForkCom) InputType() DType  { return fc.inputType }
func (fc ForkCom) OutputType() DType { return fc.outputType }

func (fc ForkCom) MeetTypes(inputType, outputType DType) Com {
	fc.inputType.Meets(inputType)
	fc.outputType.Meets(outputType)
	fc.underdefined = FanAffineTypes(&fc.outputType, &fc.inputType)
	if fc.inputType.LTE(NullType) {
		return Null
	} else if fc.outputType.Positive && len(fc.outputType.Fields) == 1 {
		for fieldName := range fc.outputType.Fields {
			return Deselect(fieldName).MeetTypes(fc.inputType, fc.outputType)
		}
		panic("Unreachable")
	} else {
		return fc
	}
}

func (fc ForkCom) Underdefined() Error {
	return fc.underdefined.Context("in fork")
}

func (fc ForkCom) Copy() Com { fc.underdefined.Remake(); return fc }

func (fc ForkCom) Invert() Com {
	return JoinCom{
		inputType:    fc.outputType,
		outputType:   fc.inputType,
		underdefined: fc.underdefined,
	}
}

func (ForkCom) TypedCom(tcb TypedComBuilder /* mutated */, inputMap, outputMap TypeMap) {
	inputMap.ForEachWith(outputMap, func(inputVar Var, outputVars []Var) {
		for _, outputVar := range outputVars {
			tcb.Equate(inputVar, outputVar)
		}
	})
}
