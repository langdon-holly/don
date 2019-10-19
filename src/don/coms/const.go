package coms

import . "don/core"

type ConstVal struct {
	P         bool /* for !struct */
	RefVal    Ref
	StructVal map[string]ConstVal
}

type constRefEntry struct {
	Output
	Val Ref
}

func putConstEntries(unitChans *[]chan<- Unit, refs *[]constRefEntry, theType DType, val ConstVal, output Output) {
	switch theType.Tag {
	case UnitTypeTag:
		if val.P {
			*unitChans = append(*unitChans, output.Unit...)
		}
	case RefTypeTag:
		if val.P {
			*refs = append(*refs, constRefEntry{Output: output, Val: val.RefVal})
		}
	case StructTypeTag:
		for fieldName, fieldType := range theType.Fields {
			putConstEntries(unitChans, refs, fieldType, val.StructVal[fieldName], output.Struct[fieldName])
		}
	}
}

type ConstCom struct {
	Type DType
	Val  ConstVal
}

func (gc ConstCom) OutputType(inputType PartialType) PartialType { return PartializeType(gc.Type) }

func (gc ConstCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	var units Output
	var refs []constRefEntry

	putConstEntries(&units.Unit, &refs, gc.Type, gc.Val, output)

	for {
		select {
		case <-input.Unit:
			units.WriteUnit()
			for _, entry := range refs {
				entry.WriteRef(entry.Val)
			}
		case <-quit:
			return
		}
	}
}
