package coms

import . "don/core"

func RunI(theType DType, input Input, output Output, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		for {
			select {
			case <-input.Unit:
				output.WriteUnit()
			case <-quit:
				return
			}
		}
	case RefTypeTag:
		for {
			select {
			case val := <-input.Ref:
				output.WriteRef(val)
			case <-quit:
				return
			}
		}
	case StructTypeTag:
		for fieldName, fieldType := range theType.Fields {
			go RunI(fieldType, input.Struct[fieldName], output.Struct[fieldName], quit)
		}
	}
}

type ICom struct{}

func (ICom) OutputType(inputType PartialType) PartialType { return inputType }

func (ICom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	RunI(inputType, input, output, quit)
}
