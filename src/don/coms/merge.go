package coms

import . "don/core"

type MergeCom struct{}

func (MergeCom) Types(inputType, outputType *DType) (bad []string, done bool) {
	if inputType.Tag == UnitTypeTag {
		bad = []string{"Unit input to merge"}
		return
	}

	inputType.RemakeFields()
	bad = FanTypes(inputType.Tag == StructTypeTag, inputType.Fields, outputType)
	if bad != nil {
		bad = append(bad, "in merge")
		return
	}
	done = inputType.Minimal()
	return
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
		go PipeUnit(output.Unit, input.Unit)
	}
}

func (MergeCom) Run(inputType, outputType DType, input Input, output Output) {
	inputs := make([]Input, len(inputType.Fields))
	i := 0
	for _, subInput := range input.Fields {
		inputs[i] = subInput
		i++
	}

	runMerge(inputs, output)
}
