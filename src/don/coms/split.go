package coms

import . "don/core"

func runSplit(theType DType, inputGetter InputGetter, outputGetters []OutputGetter, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		input := inputGetter.GetInput(theType)
		outputs := make([]Output, len(outputGetters))
		for i, outputGetter := range outputGetters {
			outputs[i] = outputGetter.GetOutput(theType)
		}

		for {
			select {
			case <-input.Unit:
				for _, output := range outputs {
					output.WriteUnit()
				}
			case <-quit:
				return
			}
		}
	case StructTypeTag:
		for fieldName, fieldType := range theType.Fields {
			subOutputGetters := make([]OutputGetter, len(outputGetters))
			for i, outputGetter := range outputGetters {
				subOutputGetters[i] = outputGetter.Struct[fieldName]
			}

			go runSplit(fieldType, inputGetter.Struct[fieldName], subOutputGetters, quit)
		}
	}
}

type SplitCom []string

func (sc SplitCom) OutputType(inputType DType) (outputType DType, impossible bool) {
	fields := make(map[string]DType, len(sc))
	for _, fieldName := range sc {
		fields[fieldName] = inputType
	}

	outputType = MakeStructType(fields)
	return
}

func (sc SplitCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	outputGetters := make([]OutputGetter, len(sc))
	for i, fieldName := range sc {
		outputGetters[i] = outputGetter.Struct[fieldName]
	}

	runSplit(inputType, inputGetter, outputGetters, quit)
}
