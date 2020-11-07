package coms

import . "don/core"

type SplitCom struct{}

func (SplitCom) Types(inputType, outputType *DType) (underdefined Error) {
	return FanAffineTypes(outputType, inputType).Context("in split")
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
	if input.Unit != nil {
		var unitChans []chan<- Unit
		for _, output := range outputs {
			if output.Unit != nil {
				unitChans = append(unitChans, output.Unit)
			}
		}
		for {
			<-input.Unit
			for _, unitChan := range unitChans {
				unitChan <- Unit{}
			}
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
