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

func runFork(input Input, outputs []Output) {
	for fieldName, subInput := range input.Fields {
		var subOutputs []Output
		for _, output := range outputs {
			if subOutput, ok := output.Fields[fieldName]; ok {
				subOutputs = append(subOutputs, subOutput)
			}
		}
		go runFork(subInput, subOutputs)
	}
	if input.Unit != nil {
		<-input.Unit
		for _, output := range outputs {
			if output.Unit != nil {
				output.Unit <- Unit{}
			}
		}
	}
}

func (ForkCom) Run(input Input, output Output) {
	outputs := make([]Output, len(output.Fields))
	i := 0
	for _, subOutput := range output.Fields {
		outputs[i] = subOutput
		i++
	}
	runFork(input, outputs)
}
