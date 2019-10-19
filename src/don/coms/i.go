package coms

import . "don/core"

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
	case RefTypeTag:
		i := input.Ref
		o := output.Ref
		for {
			select {
			case v := <-i:
				o <- v
			case <-quit:
				return
			}
		}
	case ComTypeTag:
		i := input.Com
		o := output.Com
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

type ICom struct{}

func (ICom) OutputType(inputType PartialType) PartialType { return inputType }

func (ICom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	RunI(inputType, input, output, quit)
}
