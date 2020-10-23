package coms

import . "don/core"

type MapCom struct{ Com }

func (mc MapCom) Types(inputType, outputType *DType) (bad []string, done bool) {
	done = true
	if inputType.Tag == UnitTypeTag {
		bad = []string{"Unit input to map"}
	} else if outputType.Tag == UnitTypeTag {
		bad = []string{"Unit output to map"}
	} else if inputType.Tag != UnknownTypeTag {
		if outputType.Tag == UnknownTypeTag {
			inputType.RemakeFields()
			*outputType = MakeNStructType(len(inputType.Fields))
			for fieldName, inputFieldType := range inputType.Fields {
				var outputFieldType DType
				var subDone bool
				bad, subDone = mc.Com.Types(&inputFieldType, &outputFieldType)
				if bad != nil {
					bad = append(bad, "in map field: "+fieldName)
					return
				}
				inputType.Fields[fieldName] = inputFieldType
				outputType.Fields[fieldName] = outputFieldType
				done = done && subDone
			}
		} else {
			inputType.RemakeFields()
			outputType.RemakeFields()
			for fieldName, inputFieldType := range inputType.Fields {
				outputFieldType, ok := outputType.Fields[fieldName]
				if !ok {
					bad = []string{"Fields differ in map"}
					return
				}
				var subDone bool
				bad, subDone = mc.Com.Types(&inputFieldType, &outputFieldType)
				if bad != nil {
					bad = append(bad, "in map field: "+fieldName)
					return
				}
				inputType.Fields[fieldName] = inputFieldType
				outputType.Fields[fieldName] = outputFieldType
				done = done && subDone
			}
			if len(inputType.Fields) < len(outputType.Fields) {
				bad = []string{"Fields differ in map"}
			}
		}
	} else if outputType.Tag != UnknownTypeTag {
		outputType.RemakeFields()
		*inputType = MakeNStructType(len(outputType.Fields))
		for fieldName, outputFieldType := range outputType.Fields {
			var inputFieldType DType
			var subDone bool
			bad, subDone = mc.Com.Types(&inputFieldType, &outputFieldType)
			if bad != nil {
				bad = append(bad, "in map field: "+fieldName)
				return
			}
			inputType.Fields[fieldName] = inputFieldType
			outputType.Fields[fieldName] = outputFieldType
			done = done && subDone
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
