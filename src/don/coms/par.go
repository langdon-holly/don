package coms

import "strconv"

import . "don/core"

type ParCom []Com

func (pc ParCom) Instantiate() ComInstance {
	inners := make([]ComInstance, len(pc))
	ioType := MakeNStructType(len(pc))
	for i, subCom := range pc {
		inners[i] = subCom.Instantiate()
		ioType.Fields[strconv.Itoa(i)] = UnknownType
	}
	return &parInstance{Inners: inners, inputType: ioType, outputType: ioType}
}

type parInstance struct {
	Inners                []ComInstance
	inputType, outputType DType
}

func (pi *parInstance) InputType() *DType  { return &pi.inputType }
func (pi *parInstance) OutputType() *DType { return &pi.outputType }

func (pi *parInstance) Types() {
	for i := range pi.Inners {
		idxStr := strconv.Itoa(i)
		pi.Inners[i].InputType().Meets(pi.inputType.Get(idxStr))
		pi.Inners[i].OutputType().Meets(pi.outputType.Get(idxStr))
		pi.Inners[i].Types()
		pi.inputType.Meets(pi.Inners[i].InputType().At(idxStr))
		pi.outputType.Meets(pi.Inners[i].OutputType().At(idxStr))
	}
}

func (pi parInstance) Underdefined() (underdefined Error) {
	for i, inner := range pi.Inners {
		underdefined.Ors(inner.Underdefined().Context("in " + strconv.Itoa(i) + "'th computer in par"))
	}
	return
}

func (pi parInstance) Run(input Input, output Output) {
	for i, inner := range pi.Inners {
		idxStr := strconv.Itoa(i)
		go inner.Run(input.Fields[idxStr], output.Fields[idxStr])
	}
}
