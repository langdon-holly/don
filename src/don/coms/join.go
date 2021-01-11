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

func runJoin(inputs []Input, output Output) {
	for fieldName, subOutput := range output.Fields {
		var subInputs []Input
		for _, input := range inputs {
			if subInput, ok := input.Fields[fieldName]; ok {
				subInputs = append(subInputs, subInput)
			}
		}
		go runJoin(subInputs, subOutput)
	}
	if output.Unit != nil {
		for _, input := range inputs {
			if input.Unit != nil {
				<-input.Unit
			}
		}
		output.Converge()
	}
}

func (jc JoinCom) Run(input Input, output Output) {
	inputs := make([]Input, len(jc.inputType.Fields))
	i := 0
	for _, subInput := range input.Fields {
		inputs[i] = subInput
		i++
	}
	runJoin(inputs, output)
}
