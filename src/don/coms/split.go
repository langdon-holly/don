package coms

import . "don/core"

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
	case ComTypeTag:
		i := input.Com
		a, b := outputA.Com, outputB.Com
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

type SplitCom struct{}

func (SplitCom) OutputType(inputType PartialType) PartialType {
	fields := make(map[string]PartialType, 2)
	fields["a"] = inputType
	fields["b"] = inputType
	return PartialType{P: true, Tag: StructTypeTag, Fields: fields}
}

func (SplitCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	runSplit(inputType, input, output.Struct["a"], output.Struct["b"], quit)
}
