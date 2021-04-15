package coms

import . "don/core"

func Gather() Com { return GatherCom{inputType: FieldsType} }

type GatherCom struct {
	inputType, outputType DType
	underdefined          Error
}

func (gc GatherCom) InputType() DType  { return gc.inputType }
func (gc GatherCom) OutputType() DType { return gc.outputType }

func (gc GatherCom) MeetTypes(inputType, outputType DType) Com {
	gc.inputType.Meets(inputType)
	gc.outputType.Meets(outputType)
	gc.underdefined = FanLinearTypes(&gc.inputType, &gc.outputType)
	if gc.outputType.LTE(NullType) {
		return Null
	} else if gc.inputType.Positive && len(gc.inputType.Fields) == 1 {
		for fieldName := range gc.inputType.Fields {
			return Select(fieldName).MeetTypes(gc.inputType, gc.outputType)
		}
		panic("Unreachable")
	} else {
		return gc
	}
}

func (gc GatherCom) Underdefined() Error {
	return gc.underdefined.Context("in gather")
}

func (gc GatherCom) Copy() Com { gc.underdefined.Remake(); return gc }

func (gc GatherCom) Invert() Com {
	return ScatterCom{
		inputType:    gc.outputType,
		outputType:   gc.inputType,
		underdefined: gc.underdefined,
	}
}

func (GatherCom) TypedCom(tcb TypedComBuilder /* mutated */, inputMap, outputMap TypeMap) {
	outputMap.ForEachWith(inputMap, func(outputVar Var, inputVars []Var) {
		tcb.Equate(outputVar, inputVars[0])
	})
}
