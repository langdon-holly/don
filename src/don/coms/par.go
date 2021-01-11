package coms

import "strconv"

import . "don/core"

func Par(coms []Com) Com {
	inners := make(map[int]Com, len(coms))
	inputType :=
		DType{NoUnit: true, Positive: true, Fields: make(map[string]DType)}
	outputType :=
		DType{NoUnit: true, Positive: true, Fields: make(map[string]DType)}
	for i, com := range coms {
		if _, nullp := com.(NullCom); !nullp {
			inners[i] = com
			inputType.Fields[strconv.Itoa(i)] = com.InputType()
			outputType.Fields[strconv.Itoa(i)] = com.OutputType()
		}
	}
	if len(inners) > 0 {
		return ParCom{Inners: inners, inputType: inputType, outputType: outputType}
	} else {
		return Null
	}
}

type ParCom struct {
	Inners                map[int]Com
	inputType, outputType DType
}

func (pc ParCom) InputType() DType  { return pc.inputType }
func (pc ParCom) OutputType() DType { return pc.outputType }

func (pc ParCom) MeetTypes(inputType, outputType DType) Com {
	pc.inputType.Meets(inputType)
	pc.outputType.Meets(outputType)
	for i, inner := range pc.Inners {
		idxStr := strconv.Itoa(i)
		newInputType := pc.inputType.Get(idxStr)
		newOutputType := pc.outputType.Get(idxStr)
		if !inner.InputType().LTE(newInputType) ||
			!inner.OutputType().LTE(newOutputType) {
			inner = inner.MeetTypes(newInputType, newOutputType)
			pc.inputType.Meets(inner.InputType().At(idxStr))
			pc.outputType.Meets(inner.OutputType().At(idxStr))
			if _, nullp := inner.(NullCom); nullp {
				delete(pc.Inners, i)
			} else {
				pc.Inners[i] = inner
			}
		}
	}
	if len(pc.Inners) == 0 {
		return Null
	} else if len(pc.Inners) == 1 {
		for i, inner := range pc.Inners {
			return Pipe([]Com{Select(strconv.Itoa(i)), inner, Deselect(strconv.Itoa(i))})
		}
		panic("Unreachable")
	} else {
		return pc
	}
}

func (pc ParCom) Underdefined() (underdefined Error) {
	for i, inner := range pc.Inners {
		underdefined.Ors(inner.Underdefined().Context("in " + strconv.Itoa(i) + "'th computer in par"))
	}
	return
}

func (pc ParCom) Copy() Com {
	inners := make(map[int]Com, len(pc.Inners))
	for i, inner := range pc.Inners {
		inners[i] = inner.Copy()
	}
	pc.Inners = inners
	return pc
}

func (pc ParCom) Invert() Com {
	for i, inner := range pc.Inners {
		pc.Inners[i] = inner.Invert()
	}
	pc.inputType, pc.outputType = pc.outputType, pc.inputType
	return pc
}

func (pc ParCom) Run(input Input, output Output) {
	for i, inner := range pc.Inners {
		idxStr := strconv.Itoa(i)
		go inner.Run(input.Fields[idxStr], output.Fields[idxStr])
	}
}
