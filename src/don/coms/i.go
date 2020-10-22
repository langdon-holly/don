package coms

import . "don/core"

type ICom struct{}

func PipeUnit(outputChan chan<- Unit, inputChan <-chan Unit) {
	/*if outputChan == nil {
		for {
			<-inputChan
		}
	} else */{
		for {
			<-inputChan
			outputChan <- Unit{}
		}
	}
}

func (ICom) Types(inputType, outputType *DType) (bad []string, done bool) {
	bad = MergeType2As(inputType, outputType)
	if bad != nil {
		bad = append(bad, "in unmatching I types")
		return
	}
	done = inputType.Minimal()
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
