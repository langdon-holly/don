package coms

import . "don/core"

type SplitCom struct{}

func (SplitCom) Types(inputType, outputType *DType) (bad []string, done bool) {
	if outputType.Tag == UnitTypeTag {
		bad = []string{"Unit split input"}
		return
	}

	outputType.RemakeFields()
	bad = FanTypes(outputType.Tag == StructTypeTag, outputType.Fields, inputType)
	if bad != nil {
		bad = append(bad, "in split")
		return
	}
	done = outputType.Minimal()
	return
}

func runSplit(input Input, outputs []Output) {
	for fieldName, subInput := range input.Fields {
		var subOutputs []Output
		for _, output := range outputs {
			if subOutput, ok := output.Fields[fieldName]; ok {
				subOutputs = append(subOutputs, subOutput)
			}
		}
		go runSplit(subInput, subOutputs)
	}

	for {
		<-input.Unit
		for _, output := range outputs {
			output.WriteUnit()
		}
	}
}

func (sc SplitCom) Run(inputType, outputType DType, input Input, output Output) {
	outputs := make([]Output, len(output.Fields))
	i := 0
	for _, subOutput := range output.Fields {
		outputs[i] = subOutput
		i++
	}

	runSplit(input, outputs)
}
