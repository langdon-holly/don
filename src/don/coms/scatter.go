package coms

import . "don/core"

type ScatterCom struct{}

func (ScatterCom) Instantiate() ComInstance { return &scatterInstance{} }
func (ScatterCom) Inverse() Com             { return GatherCom{} }

type scatterInstance struct {
	inputType, outputType DType
	underdefined          Error
}

func (si *scatterInstance) InputType() *DType  { return &si.inputType }
func (si *scatterInstance) OutputType() *DType { return &si.outputType }

func (si *scatterInstance) Types() {
	si.underdefined = FanLinearTypes(&si.outputType, &si.inputType)
}

func (si scatterInstance) Underdefined() Error {
	return si.underdefined.Context("in scatter")
}

func runScatter(input Input, outputs []Output) {
	for fieldName, subInput := range input.Fields {
		var subOutputs []Output
		for _, output := range outputs {
			if subOutput, ok := output.Fields[fieldName]; ok {
				subOutputs = append(subOutputs, subOutput)
			}
		}
		go runScatter(subInput, subOutputs)
	}
	if input.Unit != nil {
		var unitChan chan<- Unit
		for _, output := range outputs {
			if output.Unit != nil {
				unitChan = output.Unit
			}
		}
		PipeUnit(unitChan, input.Unit)
	}
}

func (scatterInstance) Run(input Input, output Output) {
	outputs := make([]Output, len(output.Fields))
	i := 0
	for _, subOutput := range output.Fields {
		outputs[i] = subOutput
		i++
	}
	runScatter(input, outputs)
}
