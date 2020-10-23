package coms

import . "don/core"

type ICom struct{}

func PipeUnit(outputChan chan<- Unit, inputChan <-chan Unit) {
	for {
		<-inputChan
		outputChan <- Unit{}
	}
}

func (ICom) Types(inputType, outputType *DType) (bad []string, done bool) {
	if bad = MergeType2As(inputType, outputType); bad != nil {
		bad = append(bad, "in unmatching I types")
	} else {
		done = inputType.Minimal()
	}
	return
}

func (ICom) Run(inputType, outputType DType, input Input, output Output) {
	if inputType.Tag == UnitTypeTag {
		go PipeUnit(output.Unit, input.Unit)
	}
	for fieldName, fieldType := range inputType.Fields {
		go ICom{}.Run(fieldType, fieldType, input.Fields[fieldName], output.Fields[fieldName])
	}
	return
}
