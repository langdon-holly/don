package coms

import . "don/core"

type MergeCom struct{}

func (MergeCom) Instantiate() ComInstance { return &mergeInstance{} }

type mergeInstance struct{ inputType, outputType DType }

func (mi *mergeInstance) InputType() *DType  { return &mi.inputType }
func (mi *mergeInstance) OutputType() *DType { return &mi.outputType }

func (mi *mergeInstance) Types() (underdefined Error) {
	return FanAffineTypes(&mi.inputType, &mi.outputType).Context("in merge")
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

func (mi mergeInstance) Run(input Input, output Output) {
	inputs := make([]Input, len(mi.inputType.Fields))
	i := 0
	for _, subInput := range input.Fields {
		inputs[i] = subInput
		i++
	}
	runMerge(inputs, output)
}
