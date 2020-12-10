package coms

import . "don/core"

type ChooseCom struct{}

func (ChooseCom) Instantiate() ComInstance { return &chooseInstance{} }
func (ChooseCom) Inverse() Com             { return MergeCom{} }

type chooseInstance struct {
	inputType, outputType DType
	underdefined          Error
}

func (ci *chooseInstance) InputType() *DType  { return &ci.inputType }
func (ci *chooseInstance) OutputType() *DType { return &ci.outputType }

func (ci *chooseInstance) Types() {
	ci.underdefined = FanAffineTypes(&ci.outputType, &ci.inputType)
}

func (ci chooseInstance) Underdefined() Error {
	return ci.underdefined.Context("in choose")
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
	for _, output := range outputs {
		if output.Unit != nil {
			panic("Unimplemented")
		}
	}
}

func (chooseInstance) Run(input Input, output Output) {
	outputs := make([]Output, len(output.Fields))
	i := 0
	for _, subOutput := range output.Fields {
		outputs[i] = subOutput
		i++
	}
	runChoose(input, outputs)
}
