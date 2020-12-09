package coms

import . "don/core"

type MapCom struct{ Com }

func (mc MapCom) Instantiate() ComInstance {
	return &mapInstance{Com: mc.Com, inputType: StructType, outputType: StructType}
}

type mapInstance struct {
	Com
	inputType, outputType DType
	InnerP                bool
	Inner                 ComInstance /* for InnerP */
}

func (mi *mapInstance) InputType() *DType  { return &mi.inputType }
func (mi *mapInstance) OutputType() *DType { return &mi.outputType }

func (mi *mapInstance) Types() {
	if !mi.InnerP {
		if mi.inputType.Positive {
			pipes := make([]Com, len(mi.inputType.Fields))
			i := 0
			for fieldName, _ := range mi.inputType.Fields {
				pipes[i] = PipeCom([]Com{SelectCom(fieldName), mi.Com, DeselectCom(fieldName)})
				i++
			}
			mi.InnerP = true
			mi.Inner = PipeCom([]Com{ScatterCom{}, ParCom(pipes), GatherCom{}}).Instantiate()
		} else if mi.outputType.Positive {
			pipes := make([]Com, len(mi.outputType.Fields))
			i := 0
			for fieldName, _ := range mi.outputType.Fields {
				pipes[i] = PipeCom([]Com{SelectCom(fieldName), mi.Com, DeselectCom(fieldName)})
				i++
			}
			mi.InnerP = true
			mi.Inner = PipeCom([]Com{ScatterCom{}, ParCom(pipes), GatherCom{}}).Instantiate()
		}
	}
	if mi.InnerP {
		mi.Inner.InputType().Meets(mi.inputType)
		mi.Inner.OutputType().Meets(mi.outputType)
		mi.Inner.Types()
		mi.inputType = *mi.Inner.InputType()
		mi.outputType = *mi.Inner.OutputType()
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
