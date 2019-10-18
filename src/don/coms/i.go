package coms

import . "don/core"

type ICom DType

func (com ICom) InputType() DType {
	return DType(com)
}

func (com ICom) OutputType() DType {
	return DType(com)
}

func RunI(theType DType, input Input, output Output, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		i := input.Unit
		o := output.Unit
		for {
			select {
			case <-i:
				o <- Unit{}
			case <-quit:
				return
			}
		}
	case SyntaxTypeTag:
		i := input.Syntax
		o := output.Syntax
		for {
			select {
			case v := <-i:
				o <- v
			case <-quit:
				return
			}
		}
	case GenComTypeTag:
		i := input.GenCom
		o := output.GenCom
		for {
			select {
			case v := <-i:
				o <- v
			case <-quit:
				return
			}
		}
	case StructTypeTag:
		i := input.Struct
		o := output.Struct
		for fieldName, fieldType := range theType.Fields {
			go RunI(fieldType, i[fieldName], o[fieldName], quit)
		}
	}
}

func (com ICom) Run(input Input, output Output, quit <-chan struct{}) {
	RunI(DType(com), input, output, quit)
}

type GenI struct{}

func (GenI) OutputType(inputType PartialType) PartialType { return inputType }
func (GenI) Com(inputType DType) Com                      { return ICom(inputType) }
