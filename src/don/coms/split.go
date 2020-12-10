package coms

import . "don/core"

type SplitCom struct{}

func (SplitCom) Instantiate() ComInstance { return &splitInstance{} }
func (SplitCom) Inverse() Com             { return JoinCom{} }

type splitInstance struct {
	inputType, outputType DType
	underdefined          Error
}

func (si *splitInstance) InputType() *DType  { return &si.inputType }
func (si *splitInstance) OutputType() *DType { return &si.outputType }

func (si *splitInstance) Types() {
	si.underdefined = FanAffineTypes(&si.outputType, &si.inputType)
}

func (si splitInstance) Underdefined() Error {
	return si.underdefined.Context("in split")
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

func (splitInstance) Run(input Input, output Output) {
	outputs := make([]Output, len(output.Fields))
	i := 0
	for _, subOutput := range output.Fields {
		outputs[i] = subOutput
		i++
	}
	runSplit(input, outputs)
}
