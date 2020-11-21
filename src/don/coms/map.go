package coms

import . "don/core"

type MapCom struct{ Com }

func (mc MapCom) Instantiate() ComInstance {
	return &mapInstance{Com: mc.Com, inputType: StructType, outputType: StructType}
}

type mapInstance struct {
	Com
	inputType, outputType DType
	ParP                  bool
	Par                   ComInstance /* for ParP */
}

func (mi *mapInstance) InputType() *DType  { return &mi.inputType }
func (mi *mapInstance) OutputType() *DType { return &mi.outputType }

func (mi *mapInstance) Types() {
	if !mi.ParP {
		if mi.inputType.Positive {
			pipes := make([]Com, len(mi.inputType.Fields))
			i := 0
			for fieldName, _ := range mi.inputType.Fields {
				pipes[i] = PipeCom([]Com{SelectCom(fieldName), mi.Com, DeselectCom(fieldName)})
				i++
			}
			mi.ParP = true
			mi.Par = ParCom(pipes).Instantiate()
		} else if mi.outputType.Positive {
			pipes := make([]Com, len(mi.outputType.Fields))
			i := 0
			for fieldName, _ := range mi.outputType.Fields {
				pipes[i] = PipeCom([]Com{SelectCom(fieldName), mi.Com, DeselectCom(fieldName)})
				i++
			}
			mi.ParP = true
			mi.Par = ParCom(pipes).Instantiate()
		}
	}
	if mi.ParP {
		mi.Par.InputType().Meets(mi.inputType)
		mi.Par.OutputType().Meets(mi.outputType)
		mi.Par.Types()
		mi.inputType = *mi.Par.InputType()
		mi.outputType = *mi.Par.OutputType()
	}
}

func (mi mapInstance) Underdefined() Error {
	if mi.ParP {
		return mi.Par.Underdefined().Context("in map")
	} else {
		return NewError("Negative fields in input/output to map")
	}
}

func (mi mapInstance) Run(input Input, output Output) {
	mi.Par.Run(input, output)
}
