package coms

import . "don/core"

type MergeCom struct{}

func (MergeCom) OutputType(inputType DType) (outputType DType, impossible bool) {
	if inputType.Tag == UnknownTypeTag {
		return
	}
	if inputType.Tag != StructTypeTag {
		impossible = true
		return
	}

	for _, subType := range inputType.Fields {
		outputType, impossible = MergeTypes(outputType, subType)
		if impossible {
			return
		}
	}
	return
}

func runMerge(inputTypes []DType, inputGetters []InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	if len(inputTypes) == 0 {
		return
	}

	if inputTypes[0].Tag == StructTypeTag {
		fieldInputTypeses := make(map[string][]DType)
		fieldInputGetterses := make(map[string][]InputGetter)
		for i, inputType := range inputTypes {
			inputGetter := inputGetters[i]

			for fieldName, fieldType := range inputType.Fields {
				fieldInputTypeses[fieldName] =
					append(fieldInputTypeses[fieldName], fieldType)
				fieldInputGetterses[fieldName] =
					append(fieldInputGetterses[fieldName],
						inputGetter.Struct[fieldName])
			}
		}

		for fieldName, fieldInputTypes := range fieldInputTypeses {
			fieldInputGetters := fieldInputGetterses[fieldName]
			go runMerge(fieldInputTypes, fieldInputGetters, outputGetter.Struct[fieldName], quit)
		}

		return
	}

	inputs := make([]Input, len(inputTypes))
	for i, inputType := range inputTypes {
		inputs[i] = inputGetters[i].GetInput(inputType)
	}

	output := outputGetter.GetOutput(inputTypes[0])

	for _, input := range inputs {
		go PipeUnit(output.Unit, input.Unit, quit)
	}

	return
}

func (MergeCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	inputTypes := make([]DType, len(inputType.Fields))
	inputGetters := make([]InputGetter, len(inputType.Fields))
	i := 0
	for fieldName, fieldType := range inputType.Fields {
		inputTypes[i] = fieldType
		inputGetters[i] = inputGetter.Struct[fieldName]
		i++
	}

	runMerge(inputTypes, inputGetters, outputGetter, quit)
}
