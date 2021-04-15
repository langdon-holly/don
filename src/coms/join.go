package coms

import . "don/core"

func Join() Com { return JoinCom{inputType: FieldsType} }

type JoinCom struct {
	inputType, outputType DType
	underdefined          Error
}

func (jc JoinCom) InputType() DType  { return jc.inputType }
func (jc JoinCom) OutputType() DType { return jc.outputType }

func (jc JoinCom) MeetTypes(inputType, outputType DType) Com {
	jc.inputType.Meets(inputType)
	jc.outputType.Meets(outputType)
	jc.underdefined = FanAffineTypes(&jc.inputType, &jc.outputType)
	if jc.outputType.LTE(NullType) {
		return Null
	} else if jc.inputType.Positive && len(jc.inputType.Fields) == 1 {
		for fieldName := range jc.inputType.Fields {
			return Select(fieldName).MeetTypes(jc.inputType, jc.outputType)
		}
		panic("Unreachable")
	} else {
		return jc
	}
}

func (jc JoinCom) Underdefined() Error {
	return jc.underdefined.Context("in join")
}

func (jc JoinCom) Copy() Com { jc.underdefined.Remake(); return jc }

func (jc JoinCom) Invert() Com {
	return ForkCom{
		inputType:    jc.outputType,
		outputType:   jc.inputType,
		underdefined: jc.underdefined,
	}
}

func (JoinCom) TypedCom(tcb TypedComBuilder /* mutated */, inputMap, outputMap TypeMap) {
	outputMap.ForEachWith(inputMap, func(outputVar Var, inputVars []Var) {
		for _, inputVar := range inputVars {
			tcb.Equate(outputVar, inputVar)
		}
	})
}
