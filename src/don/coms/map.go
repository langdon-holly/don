package coms

import . "don/core"

type MapCom struct{ Com }

func (mc MapCom) Types(inputType, outputType *DType) (done bool) {
	inputType.Meets(StructType)
	outputType.Meets(StructType)
	if inputType.Positive {
		inputType.RemakeFields()
		if outputType.Positive {
			outputType.RemakeFields()
		} else {
			*outputType = MakeNStructType(len(inputType.Fields))
			for fieldName := range inputType.Fields {
				outputType.Fields[fieldName] = UnknownType
			}
		}
	} else if outputType.Positive {
		outputType.RemakeFields()
		*inputType = MakeNStructType(len(outputType.Fields))
		for fieldName := range outputType.Fields {
			inputType.Fields[fieldName] = UnknownType
		}
	}
	if inputType.Positive {
		done = true
		for fieldName, inputFieldType := range inputType.Fields {
			outputFieldType := outputType.Get(fieldName)
			done = done && mc.Com.Types(&inputFieldType, &outputFieldType)

			inputType.Fields[fieldName] = inputFieldType
			outputType.Fields[fieldName] = outputFieldType
			if inputFieldType.LTE(NullType) {
				delete(inputType.Fields, fieldName)
				delete(outputType.Fields, fieldName)
			}
		}
		for fieldName := range outputType.Fields {
			if _, ok := inputType.Fields[fieldName]; !ok {
				delete(outputType.Fields, fieldName)
			}
		}
	}
	return
}

func (mc MapCom) Run(inputType, outputType DType, input Input, output Output) {
	pipes := make([]Com, len(inputType.Fields))
	i := 0
	for fieldName, _ := range inputType.Fields {
		pipes[i] = PipeCom([]Com{SelectCom(fieldName), mc.Com, DeselectCom(fieldName)})
		i++
	}
	SplitMergeCom(pipes).Run(inputType, outputType, input, output)
}
