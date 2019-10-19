package coms

import . "don/core"

func listenUnit(input <-chan Unit, output chan<- Unit, quit <-chan struct{}) {
	for {
		select {
		case <-input:
			output <- Unit{}
		case <-quit:
			return
		}
	}
}

func listenRef(input <-chan Ref, output chan<- Ref, quit <-chan struct{}) {
	for {
		select {
		case val := <-input:
			output <- val
		case <-quit:
			return
		}
	}
}

func runMerge(theType DType, inputs []Input, output Output, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		for _, input := range inputs {
			go listenUnit(input.Unit, output.Unit, quit)
		}
	case RefTypeTag:
		for _, input := range inputs {
			go listenRef(input.Ref, output.Ref, quit)
		}
	case StructTypeTag:
		for fieldName, fieldType := range theType.Fields {
			subInputs := make([]Input, len(inputs))
			for i, input := range inputs {
				subInputs[i] = input.Struct[fieldName]
			}

			go runMerge(fieldType, subInputs, output.Struct[fieldName], quit)
		}
	}
}

type MergeCom struct{}

func (MergeCom) OutputType(inputType PartialType) (ret PartialType) {
	if inputType.P {
		for _, subType := range inputType.Fields {
			ret = MergePartialTypes(ret, subType)
		}
	}
	return
}

func (MergeCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	var elementType DType
	inputs := make([]Input, len(inputType.Fields))
	i := 0
	for fieldName, subType := range inputType.Fields {
		elementType = subType
		inputs[i] = input.Struct[fieldName]
		i++
	}

	runMerge(elementType, inputs, output, quit)
}
