package coms

import . "don/core"

type SplitCom DType

func (com SplitCom) InputType() DType {
	return DType(com)
}

func (com SplitCom) OutputType() DType {
	theType := DType(com)
	fields := make(map[string]DType, 2)
	fields["a"] = theType
	fields["b"] = theType
	return DType{StructTypeTag, fields}
}

func runSplit(theType DType, input Input, outputA, outputB Output, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		i := input.Unit
		a, b := outputA.Unit, outputB.Unit
		for {
			select {
			case <-i:
				a <- Unit{}
				b <- Unit{}
			case <-quit:
				return
			}
		}
	case SyntaxTypeTag:
		i := input.Syntax
		a, b := outputA.Syntax, outputB.Syntax
		for {
			select {
			case v := <-i:
				a <- v
				b <- v
			case <-quit:
				return
			}
		}
	case GenComTypeTag:
		i := input.GenCom
		a, b := outputA.GenCom, outputB.GenCom
		for {
			select {
			case v := <-i:
				a <- v
				b <- v
			case <-quit:
				return
			}
		}
	case StructTypeTag:
		i := input.Struct
		a, b := outputA.Struct, outputB.Struct
		for fieldName, fieldType := range theType.Fields {
			go runSplit(fieldType, i[fieldName], a[fieldName], b[fieldName], quit)
		}
	}
}

func (com SplitCom) Run(input Input, output Output, quit <-chan struct{}) {
	outputStruct := output.Struct
	runSplit(DType(com), input, outputStruct["a"], outputStruct["b"], quit)
}

type GenSplit struct{}

func (GenSplit) OutputType(inputType PartialType) PartialType {
	fields := make(map[string]PartialType, 2)
	fields["a"] = inputType
	fields["b"] = inputType
	return PartialType{P: true, Tag: StructTypeTag, Fields: fields}
}

func (GenSplit) Com(inputType DType) Com { return SplitCom(inputType) }
