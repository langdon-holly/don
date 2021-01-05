package coms

import . "don/core"

type ForkCom struct{}

func (ForkCom) Instantiate() ComInstance { return &forkInstance{} }
func (ForkCom) Inverse() Com             { return JoinCom{} }

type forkInstance struct {
	inputType, outputType DType
	underdefined          Error
}

func (fi *forkInstance) InputType() *DType  { return &fi.inputType }
func (fi *forkInstance) OutputType() *DType { return &fi.outputType }

func (fi *forkInstance) Types() {
	fi.underdefined = FanAffineTypes(&fi.outputType, &fi.inputType)
}

func (fi forkInstance) Underdefined() Error {
	return fi.underdefined.Context("in fork")
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

func (forkInstance) Run(input Input, output Output) {
	outputs := make([]Output, len(output.Fields))
	i := 0
	for _, subOutput := range output.Fields {
		outputs[i] = subOutput
		i++
	}
	runFork(input, outputs)
}
