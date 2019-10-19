package coms

import . "don/core"

func runSplit(theType DType, input Input, outputs []Output, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		for {
			select {
			case <-input.Unit:
				for _, output := range outputs {
					output.Unit <- Unit{}
				}
			case <-quit:
				return
			}
		}
	case RefTypeTag:
		for {
			select {
			case val := <-input.Ref:
				for _, output := range outputs {
					output.Ref <- val
				}
			case <-quit:
				return
			}
		}
	case SyntaxTypeTag:
		for {
			select {
			case val := <-input.Syntax:
				for _, output := range outputs {
					output.Syntax <- val
				}
			case <-quit:
				return
			}
		}
	case ComTypeTag:
		for {
			select {
			case val := <-input.Com:
				for _, output := range outputs {
					output.Com <- val
				}
			case <-quit:
				return
			}
		}
	case StructTypeTag:
		for fieldName, fieldType := range theType.Fields {
			subOutputs := make([]Output, len(outputs))
			for i, output := range outputs {
				subOutputs[i] = output.Struct[fieldName]
			}

			go runSplit(fieldType, input.Struct[fieldName], subOutputs, quit)
		}
	}
}

type SplitCom []string

func (sc SplitCom) OutputType(inputType PartialType) PartialType {
	fields := make(map[string]PartialType, len(sc))
	for _, fieldName := range sc {
		fields[fieldName] = inputType
	}

	return PartialType{P: true, Tag: StructTypeTag, Fields: fields}
}

func (sc SplitCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	outputs := make([]Output, len(sc))
	for i, fieldName := range sc {
		outputs[i] = output.Struct[fieldName]
	}

	runSplit(inputType, input, outputs, quit)
}
