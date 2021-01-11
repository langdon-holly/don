package coms

import . "don/core"

func Map(com Com) Com {
	return MapCom{Com: com, inputType: FieldsType, outputType: FieldsType}
}

type MapCom struct {
	Com                   Com
	inputType, outputType DType
}

func (mc MapCom) InputType() DType  { return mc.inputType }
func (mc MapCom) OutputType() DType { return mc.outputType }

func (mc MapCom) MeetTypes(inputType, outputType DType) Com {
	mc.inputType.Meets(inputType)
	mc.outputType.Meets(outputType)
	if mc.inputType.Positive || mc.outputType.Positive {
		fieldNames := mc.outputType.Fields
		if mc.inputType.Positive {
			fieldNames = mc.inputType.Fields
		}
		pipes := make([]Com, len(fieldNames))
		i := 0
		for fieldName := range fieldNames {
			pipes[i] = Pipe([]Com{Select(fieldName), mc.Com.Copy(), Deselect(fieldName)})
			i++
		}
		inner := Pipe([]Com{Scatter(), Par(pipes), Gather()})
		return inner.MeetTypes(mc.inputType, mc.outputType)
	} else {
		return mc
	}
}

func (mc MapCom) Underdefined() Error {
	return NewError("Negative fields in input/output to map")
}

func (mc MapCom) Copy() Com { mc.Com = mc.Com.Copy(); return mc }

func (mc MapCom) Invert() Com {
	mc.Com = mc.Com.Invert()
	mc.inputType, mc.outputType = mc.outputType, mc.inputType
	return mc
}

func (mc MapCom) Run(input Input, output Output) {
	panic("Negative fields in input/output to map")
}
