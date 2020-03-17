package coms

import . "don/core"

type MergeCom struct{}

func (MergeCom) OutputType(inputType DType) DType {
	if inputType.Lvl != NormalTypeLvl {
		return inputType
	}
	if inputType.Tag != StructTypeTag {
		return ImpossibleType
	}

	ret := UnknownType
	for _, subType := range inputType.Fields {
		ret = MergeTypes(ret, subType)
	}
	return ret
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

	if inputTypes[0].Tag == UnitTypeTag {
		for _, input := range inputs {
			go PipeUnit(output.Unit, input.Unit, quit)
		}
	} else { //inputTypes[0].Tag == RefTypeTag
		for _, input := range inputs {
			go PipeRef(output.Ref, input.Ref, quit)
		}
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
