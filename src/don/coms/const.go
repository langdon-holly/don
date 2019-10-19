package coms

import . "don/core"

type ConstVal struct {
	P         bool /* for !struct */
	RefVal    Ref
	StructVal map[string]ConstVal
}

type constRefEntry struct {
	Chan chan<- Ref
	Val  Ref
}

func putConstEntries(units *[]chan<- Unit, refs *[]constRefEntry, theType DType, val ConstVal, output Output) {
	switch theType.Tag {
	case UnitTypeTag:
		if val.P {
			*units = append(*units, output.Unit)
		}
	case RefTypeTag:
		if val.P {
			*refs = append(*refs, constRefEntry{output.Ref, val.RefVal})
		}
	case StructTypeTag:
		for fieldName, fieldType := range theType.Fields {
			putConstEntries(units, refs, fieldType, val.StructVal[fieldName], output.Struct[fieldName])
		}
	}
}

type ConstCom struct {
	Type DType
	Val  ConstVal
}

func (gc ConstCom) OutputType(inputType PartialType) PartialType { return PartializeType(gc.Type) }

func (gc ConstCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	var units []chan<- Unit
	var refs []constRefEntry

	putConstEntries(&units, &refs, gc.Type, gc.Val, output)

	for {
		select {
		case <-input.Unit:
			for _, entry := range units {
				entry <- Unit{}
			}
			for _, entry := range refs {
				entry.Chan <- entry.Val
			}
		case <-quit:
			return
		}
	}
}
