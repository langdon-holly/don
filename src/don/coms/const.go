package coms

import . "don/core"

type ConstVal struct {
	P   bool
	Val interface{}
}

type ConstCom struct {
	Type DType
	Val  interface{}
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

func putConstEntries(units *[]chan<- Unit, syntaxen *[]constSyntaxEntry, genComs *[]constGenComEntry, theType DType, val interface{}, output interface{}) {
	switch theType.Tag {
	case UnitTypeTag:
		constVal := val.(ConstVal)
		if constVal.P {
			*units = append(*units, output.(chan<- Unit))
		}
	case SyntaxTypeTag:
		constVal := val.(ConstVal)
		if constVal.P {
			*syntaxen = append(*syntaxen, constSyntaxEntry{output.(chan<- Syntax), constVal.Val.(Syntax)})
		}
	case GenComTypeTag:
		constVal := val.(ConstVal)
		if constVal.P {
			*genComs = append(*genComs, constGenComEntry{output.(chan<- GenCom), constVal.Val.(GenCom)})
		}
	case StructTypeTag:
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			putConstEntries(units, syntaxen, genComs, fieldType, val.(Struct)[fieldName], output.(Struct)[fieldName])
		}
	}
}

func (com ConstCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	i := input.(<-chan Unit)

	var units []chan<- Unit
	var syntaxen []constSyntaxEntry
	var genComs []constGenComEntry

	putConstEntries(&units, &syntaxen, &genComs, com.Type, com.Val, output)

	for {
		select {
		case <-i:
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
	Val  interface{}
}

func (gc GenConst) OutputType(inputType PartialType) PartialType { return PartializeType(gc.Type) }
func (gc GenConst) Com(inputType DType) Com                      { return ConstCom{Type: gc.Type, Val: gc.Val} }
