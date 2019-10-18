package coms

import . "don/core"

type ConstVal struct {
	P         bool /* for !struct */
	SyntaxVal Syntax
	GenComVal GenCom
	StructVal map[string]ConstVal
}

type ConstCom struct {
	Type DType
	Val  ConstVal
}

func (com ConstCom) InputType() DType {
	return UnitType
}

func (com ConstCom) OutputType() DType {
	return com.Type
}

type constSyntaxEntry struct {
	Chan chan<- Syntax
	Val  Syntax
}

type constGenComEntry struct {
	Chan chan<- GenCom
	Val  GenCom
}

func putConstEntries(units *[]chan<- Unit, syntaxen *[]constSyntaxEntry, genComs *[]constGenComEntry, theType DType, val ConstVal, output Output) {
	switch theType.Tag {
	case UnitTypeTag:
		if val.P {
			*units = append(*units, output.Unit)
		}
	case SyntaxTypeTag:
		if val.P {
			*syntaxen = append(*syntaxen, constSyntaxEntry{output.Syntax, val.SyntaxVal})
		}
	case GenComTypeTag:
		if val.P {
			*genComs = append(*genComs, constGenComEntry{output.GenCom, val.GenComVal})
		}
	case StructTypeTag:
		for fieldName, fieldType := range theType.Fields {
			putConstEntries(units, syntaxen, genComs, fieldType, val.StructVal[fieldName], output.Struct[fieldName])
		}
	}
}

func (com ConstCom) Run(input Input, output Output, quit <-chan struct{}) {
	var units []chan<- Unit
	var syntaxen []constSyntaxEntry
	var genComs []constGenComEntry

	putConstEntries(&units, &syntaxen, &genComs, com.Type, com.Val, output)

	for {
		select {
		case <-input.Unit:
			for _, entry := range units {
				entry <- Unit{}
			}
			for _, entry := range syntaxen {
				entry.Chan <- entry.Val
			}
			for _, entry := range genComs {
				entry.Chan <- entry.Val
			}
		case <-quit:
			return
		}
	}
}

type GenConst struct {
	Type DType
	Val  ConstVal
}

func (gc GenConst) OutputType(inputType PartialType) PartialType { return PartializeType(gc.Type) }
func (gc GenConst) Com(inputType DType) Com                      { return ConstCom{Type: gc.Type, Val: gc.Val} }
