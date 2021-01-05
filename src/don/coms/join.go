package coms

import . "don/core"

type JoinCom struct{}

func (JoinCom) Instantiate() ComInstance { return &joinInstance{} }
func (JoinCom) Inverse() Com             { return ForkCom{} }

type joinInstance struct {
	inputType, outputType DType
	underdefined          Error
}

func (ji *joinInstance) InputType() *DType  { return &ji.inputType }
func (ji *joinInstance) OutputType() *DType { return &ji.outputType }

func (ji *joinInstance) Types() {
	ji.underdefined = FanAffineTypes(&ji.inputType, &ji.outputType)
}

func (ji joinInstance) Underdefined() Error {
	return ji.underdefined.Context("in join")
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

func (ji joinInstance) Run(input Input, output Output) {
	inputs := make([]Input, len(ji.inputType.Fields))
	i := 0
	for _, subInput := range input.Fields {
		inputs[i] = subInput
		i++
	}
	runJoin(inputs, output)
}
