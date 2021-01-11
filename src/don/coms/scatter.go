package coms

import . "don/core"

func Scatter() Com { return ScatterCom{outputType: FieldsType} }

type ScatterCom struct {
	inputType, outputType DType
	underdefined          Error
}

func (sc ScatterCom) InputType() DType  { return sc.inputType }
func (sc ScatterCom) OutputType() DType { return sc.outputType }

func (sc ScatterCom) MeetTypes(inputType, outputType DType) Com {
	sc.inputType.Meets(inputType)
	sc.outputType.Meets(outputType)
	sc.underdefined = FanLinearTypes(&sc.outputType, &sc.inputType)
	if sc.inputType.LTE(NullType) {
		return Null
	} else if sc.outputType.Positive && len(sc.outputType.Fields) == 1 {
		for fieldName := range sc.outputType.Fields {
			return Deselect(fieldName).MeetTypes(sc.inputType, sc.outputType)
		}
		panic("Unreachable")
	} else {
		return sc
	}
}

func (sc ScatterCom) Underdefined() Error {
	return sc.underdefined.Context("in scatter")
}

func (sc ScatterCom) Copy() Com { sc.underdefined.Remake(); return sc }

func (sc ScatterCom) Invert() Com {
	return GatherCom{
		inputType:    sc.outputType,
		outputType:   sc.inputType,
		underdefined: sc.underdefined,
	}
}

func runScatter(input Input, outputs []Output) {
	for fieldName, subInput := range input.Fields {
		var subOutputs []Output
		for _, output := range outputs {
			if subOutput, ok := output.Fields[fieldName]; ok {
				subOutputs = append(subOutputs, subOutput)
			}
		}
		go runScatter(subInput, subOutputs)
	}
	if input.Unit != nil {
		var unitChan chan<- Unit
		for _, output := range outputs {
			if output.Unit != nil {
				unitChan = output.Unit
			}
		}
		PipeUnit(unitChan, input.Unit)
	}
}

func (ScatterCom) Run(input Input, output Output) {
	outputs := make([]Output, len(output.Fields))
	i := 0
	for _, subOutput := range output.Fields {
		outputs[i] = subOutput
		i++
	}
	runScatter(input, outputs)
}
