package coms

import . "don/core"

type GatherCom struct{}

func (GatherCom) Instantiate() ComInstance { return &gatherInstance{} }
func (GatherCom) Inverse() Com             { return ScatterCom{} }

type gatherInstance struct {
	inputType, outputType DType
	underdefined          Error
}

func (gi *gatherInstance) InputType() *DType  { return &gi.inputType }
func (gi *gatherInstance) OutputType() *DType { return &gi.outputType }

func (gi *gatherInstance) Types() {
	gi.underdefined = FanLinearTypes(&gi.inputType, &gi.outputType)
}

func (gi gatherInstance) Underdefined() Error {
	return gi.underdefined.Context("in gather")
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

func (gi gatherInstance) Run(input Input, output Output) {
	inputs := make([]Input, len(gi.inputType.Fields))
	i := 0
	for _, subInput := range input.Fields {
		inputs[i] = subInput
		i++
	}
	runGather(inputs, output)
}
