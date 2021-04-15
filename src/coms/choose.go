package coms

import . "don/core"

func Choose() Com { return ChooseCom{outputType: FieldsType} }

type ChooseCom struct {
	inputType, outputType DType
	underdefined          Error
}

func (cc ChooseCom) InputType() DType  { return cc.inputType }
func (cc ChooseCom) OutputType() DType { return cc.outputType }

func (cc ChooseCom) MeetTypes(inputType, outputType DType) Com {
	cc.inputType.Meets(inputType)
	cc.outputType.Meets(outputType)
	cc.underdefined = FanAffineTypes(&cc.outputType, &cc.inputType)
	if cc.inputType.LTE(NullType) {
		return Null
	} else if cc.outputType.Positive && len(cc.outputType.Fields) == 1 {
		for fieldName := range cc.outputType.Fields {
			return Deselect(fieldName).MeetTypes(cc.inputType, cc.outputType)
		}
		panic("Unreachable")
	} else {
		return cc
	}
}

func (cc ChooseCom) Underdefined() Error {
	return cc.underdefined.Context("in choose")
}

func (cc ChooseCom) Copy() Com { cc.underdefined.Remake(); return cc }

func (cc ChooseCom) Invert() Com {
	return MergeCom{
		inputType:    cc.outputType,
		outputType:   cc.inputType,
		underdefined: cc.underdefined,
	}
}

func (ChooseCom) TypedCom(tcb TypedComBuilder /* mutated */, inputMap, outputMap TypeMap) {
	inputMap.ForEachWith(outputMap, func(inputVar Var, outputVars []Var) {
		if len(outputVars) == 1 {
			tcb.Equate(inputVar, outputVars[0])
		} else {
			tcb.Add(&ChooseNode{In: inputVar, Out: outputVars})
		}
	})
}
