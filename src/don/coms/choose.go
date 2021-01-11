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

func runChoose(input Input, outputs []Output) {
	for fieldName, subInput := range input.Fields {
		var subOutputs []Output
		for _, output := range outputs {
			if subOutput, ok := output.Fields[fieldName]; ok {
				subOutputs = append(subOutputs, subOutput)
			}
		}
		go runChoose(subInput, subOutputs)
	}
	var onlyOutput chan<- Unit
	for _, output := range outputs {
		if output.Unit != nil {
			if onlyOutput != nil {
				panic("Unimplemented")
			} else if onlyOutput = output.Unit; true {
			}
		}
	}
	if onlyOutput != nil {
		PipeUnit(onlyOutput, input.Unit)
	}
}

func (ChooseCom) Run(input Input, output Output) {
	outputs := make([]Output, len(output.Fields))
	i := 0
	for _, subOutput := range output.Fields {
		outputs[i] = subOutput
		i++
	}
	runChoose(input, outputs)
}
