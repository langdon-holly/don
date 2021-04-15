package coms

import . "don/core"

func Merge() Com { return MergeCom{inputType: FieldsType} }

type MergeCom struct {
	inputType, outputType DType
	underdefined          Error
}

func (mc MergeCom) InputType() DType  { return mc.inputType }
func (mc MergeCom) OutputType() DType { return mc.outputType }

func (mc MergeCom) MeetTypes(inputType, outputType DType) Com {
	mc.inputType.Meets(inputType)
	mc.outputType.Meets(outputType)
	mc.underdefined = FanAffineTypes(&mc.inputType, &mc.outputType)
	if mc.outputType.LTE(NullType) {
		return Null
	} else if mc.inputType.Positive && len(mc.inputType.Fields) == 1 {
		for fieldName := range mc.inputType.Fields {
			return Select(fieldName).MeetTypes(mc.inputType, mc.outputType)
		}
		panic("Unreachable")
	} else {
		return mc
	}
}

func (mc MergeCom) Underdefined() Error {
	return mc.underdefined.Context("in merge")
}

func (mc MergeCom) Copy() Com { mc.underdefined.Remake(); return mc }

func (mc MergeCom) Invert() Com {
	return ChooseCom{
		inputType:    mc.outputType,
		outputType:   mc.inputType,
		underdefined: mc.underdefined,
	}
}

func (MergeCom) TypedCom(tcb TypedComBuilder /* mutated */, inputMap, outputMap TypeMap) {
	outputMap.ForEachWith(inputMap, func(outputVar Var, inputVars []Var) {
		if len(inputVars) == 1 {
			tcb.Equate(outputVar, inputVars[0])
		} else {
			tcb.Add(&MergeNode{In: inputVars, Out: outputVar})
		}
	})
}
