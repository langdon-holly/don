package coms

import . "don/core"

func Gather() Com { return &GatherCom{inputType: FieldsType} }

type GatherCom struct {
	inputType, outputType DType
	underdefined          Error
}

func (gc *GatherCom) InputType() *DType  { return &gc.inputType }
func (gc *GatherCom) OutputType() *DType { return &gc.outputType }

func (gc *GatherCom) Types() Com {
	gc.underdefined = FanLinearTypes(&gc.inputType, &gc.outputType)
	if gc.outputType.LTE(NullType) {
		return Null
	} else if gc.inputType.Positive && len(gc.inputType.Fields) == 1 {
		for fieldName := range gc.inputType.Fields {
			sc := Select(fieldName)
			sc.InputType().Meets(gc.inputType)
			sc.OutputType().Meets(gc.outputType)
			return sc.Types()
		}
		panic("Unreachable")
	} else {
		return gc
	}
}

func (gc GatherCom) Underdefined() Error {
	return gc.underdefined.Context("in gather")
}

func (gc GatherCom) Copy() Com { gc.underdefined.Remake(); return &gc }

func (gc GatherCom) Invert() Com {
	return &ScatterCom{
		inputType:    gc.outputType,
		outputType:   gc.inputType,
		underdefined: gc.underdefined,
	}
}

func runGather(inputs []Input, output Output) {
	for fieldName, subOutput := range output.Fields {
		var subInputs []Input
		for _, input := range inputs {
			if subInput, ok := input.Fields[fieldName]; ok {
				subInputs = append(subInputs, subInput)
			}
		}
		go runGather(subInputs, subOutput)
	}
	if output.Unit != nil {
		var unitChan <-chan Unit
		for _, input := range inputs {
			if input.Unit != nil {
				unitChan = input.Unit
			}
		}
		PipeUnit(output.Unit, unitChan)
	}
}

func (gc GatherCom) Run(input Input, output Output) {
	inputs := make([]Input, len(gc.inputType.Fields))
	i := 0
	for _, subInput := range input.Fields {
		inputs[i] = subInput
		i++
	}
	runGather(inputs, output)
}
