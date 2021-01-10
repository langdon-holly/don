package coms

import . "don/core"

func Merge() Com { return &MergeCom{inputType: FieldsType} }

type MergeCom struct {
	inputType, outputType DType
	underdefined          Error
}

func (mc *MergeCom) InputType() *DType  { return &mc.inputType }
func (mc *MergeCom) OutputType() *DType { return &mc.outputType }

func (mc *MergeCom) Types() Com {
	mc.underdefined = FanAffineTypes(&mc.inputType, &mc.outputType)
	if mc.outputType.LTE(NullType) {
		return Null
	} else if mc.inputType.Positive && len(mc.inputType.Fields) == 1 {
		for fieldName := range mc.inputType.Fields {
			sc := Select(fieldName)
			sc.InputType().Meets(mc.inputType)
			sc.OutputType().Meets(mc.outputType)
			return sc.Types()
		}
		panic("Unreachable")
	} else {
		return mc
	}
}

func (mc MergeCom) Underdefined() Error {
	return mc.underdefined.Context("in merge")
}

func (mc MergeCom) Copy() Com { mc.underdefined.Remake(); return &mc }

func (mc MergeCom) Invert() Com {
	return &ChooseCom{
		inputType:    mc.outputType,
		outputType:   mc.inputType,
		underdefined: mc.underdefined,
	}
}

func runMerge(inputs []Input, output Output) {
	for fieldName, subOutput := range output.Fields {
		var subInputs []Input
		for _, input := range inputs {
			if subInput, ok := input.Fields[fieldName]; ok {
				subInputs = append(subInputs, subInput)
			}
		}
		go runMerge(subInputs, subOutput)
	}
	for _, input := range inputs {
		if input.Unit != nil {
			go PipeUnit(output.Unit, input.Unit)
		}
	}
}

func (mc MergeCom) Run(input Input, output Output) {
	inputs := make([]Input, len(mc.inputType.Fields))
	i := 0
	for _, subInput := range input.Fields {
		inputs[i] = subInput
		i++
	}
	runMerge(inputs, output)
}
