package coms

import . "don/core"

type MapCom struct{ Com Com }

func (mc MapCom) Instantiate() ComInstance {
	return &mapInstance{Com: mc.Com, inputType: FieldsType, outputType: FieldsType}
}

func (mc MapCom) Inverse() Com { return MapCom{Com: mc.Com.Inverse()} }

type mapInstance struct {
	InnerP bool

	// for !InnerP
	Com                   Com
	inputType, outputType DType

	// for InnerP
	Inner ComInstance
}

func (mi *mapInstance) InputType() *DType {
	if mi.InnerP {
		return mi.Inner.InputType()
	} else {
		return &mi.inputType
	}
}
func (mi *mapInstance) OutputType() *DType {
	if mi.InnerP {
		return mi.Inner.OutputType()
	} else {
		return &mi.outputType
	}
}

func (mi *mapInstance) Types() {
	if !mi.InnerP && mi.inputType.Positive || mi.outputType.Positive {
		fieldNames := mi.outputType.Fields
		if mi.inputType.Positive {
			fieldNames = mi.inputType.Fields
		}
		pipes := make([]Com, len(fieldNames))
		i := 0
		for fieldName := range fieldNames {
			pipes[i] = PipeCom([]Com{SelectCom(fieldName), mi.Com, DeselectCom(fieldName)})
			i++
		}
		inner := PipeCom(
			[]Com{ScatterCom{}, ParCom(pipes), GatherCom{}}).Instantiate()
		inner.InputType().Meets(mi.inputType)
		inner.OutputType().Meets(mi.outputType)
		*mi = mapInstance{InnerP: true, Inner: inner}
	}
	if mi.InnerP {
		mi.Inner.Types()
	}
}

func (mi mapInstance) Underdefined() Error {
	if mi.InnerP {
		return mi.Inner.Underdefined().Context("in map")
	} else {
		return NewError("Negative fields in input/output to map")
	}
}

func (mi mapInstance) Run(input Input, output Output) {
	mi.Inner.Run(input, output)
}
