package coms

import "strconv"

import . "don/core"

type ParCom []Com

const (
	parComSubComIdInnerTag = iota
	parComSubComIdFanOutTag
	parComSubComIdFanInTag
)

type parComSubComId struct {
	Tag int
	Idx int /* for Tag == parComSubComIdInnerTag */
}

func (pc ParCom) Instantiate() ComInstance {
	inners := make([]ComInstance, len(pc))
	toType := make(map[parComSubComId]struct{}, len(pc)+2)
	for i, subCom := range pc {
		inners[i] = subCom.Instantiate()
		toType[parComSubComId{Idx: i}] = struct{}{}
	}
	return &parInstance{Inners: inners, ToType: toType}
}

type parInstance struct {
	Inners                                []ComInstance
	ToType                                map[parComSubComId]struct{}
	inputType, outputType                 DType
	FanOutUnderdefined, FanInUnderdefined Error
}

func (pi *parInstance) InputType() *DType  { return &pi.inputType }
func (pi *parInstance) OutputType() *DType { return &pi.outputType }

func (pi *parInstance) Types() {
	pi.ToType[parComSubComId{Tag: parComSubComIdFanOutTag}] = struct{}{}
	pi.ToType[parComSubComId{Tag: parComSubComIdFanInTag}] = struct{}{}
	for len(pi.ToType) > 0 {
		var typeNext parComSubComId
		for typeNext = range pi.ToType {
			delete(pi.ToType, typeNext)
			break
		}
		switch typeNext.Tag {
		case parComSubComIdInnerTag:
			innerInputTypeBefore := *pi.Inners[typeNext.Idx].InputType()
			innerOutputTypeBefore := *pi.Inners[typeNext.Idx].OutputType()
			pi.Inners[typeNext.Idx].Types()
			if !innerInputTypeBefore.LTE(*pi.Inners[typeNext.Idx].InputType()) {
				pi.ToType[parComSubComId{Tag: parComSubComIdFanOutTag}] = struct{}{}
			}
			if !innerOutputTypeBefore.LTE(*pi.Inners[typeNext.Idx].OutputType()) {
				pi.ToType[parComSubComId{Tag: parComSubComIdFanInTag}] = struct{}{}
			}
		case parComSubComIdFanOutTag:
			innerInputTypes := make([]DType, len(pi.Inners))
			for i, inner := range pi.Inners {
				innerInputTypes[i] = *inner.InputType()
			}
			innerInputTypesBefore := append([]DType{}, innerInputTypes...)
			pi.FanOutUnderdefined = FanLinearTypes(innerInputTypes, &pi.inputType)
			for i, innerInputType := range innerInputTypes {
				if !innerInputTypesBefore[i].LTE(innerInputType) {
					*pi.Inners[i].InputType() = innerInputType
					pi.ToType[parComSubComId{Idx: i}] = struct{}{}
				}
			}
		case parComSubComIdFanInTag:
			innerOutputTypes := make([]DType, len(pi.Inners))
			for i, inner := range pi.Inners {
				innerOutputTypes[i] = *inner.OutputType()
			}
			innerOutputTypesBefore := append([]DType{}, innerOutputTypes...)
			pi.FanInUnderdefined = FanLinearTypes(innerOutputTypes, &pi.outputType)
			for i, innerOutputType := range innerOutputTypes {
				if !innerOutputTypesBefore[i].LTE(innerOutputType) {
					*pi.Inners[i].OutputType() = innerOutputType
					pi.ToType[parComSubComId{Idx: i}] = struct{}{}
				}
			}
		}
	}
}

func (pi parInstance) Underdefined() (underdefined Error) {
	underdefined.Ors(pi.FanOutUnderdefined.Context("in par fan-out"))
	underdefined.Ors(pi.FanInUnderdefined.Context("in par fan-in"))
	for i, inner := range pi.Inners {
		underdefined.Ors(inner.Underdefined().Context("in " + strconv.Itoa(i) + "'th computer in par"))
	}
	return
}

// inputType.Underdefined() == nil
func subInput(input Input, inputType DType) (sub Input) {
	if !inputType.NoUnit {
		sub.Unit = input.Unit
	}
	sub.Fields = make(map[string]Input, len(inputType.Fields))
	for fieldName, fieldType := range inputType.Fields {
		sub.Fields[fieldName] = subInput(input.Fields[fieldName], fieldType)
	}
	return
}

// outputType.Underdefined() == nil
func subOutput(output Output, outputType DType) (sub Output) {
	if !outputType.NoUnit {
		sub.Unit = output.Unit
	}
	sub.Fields = make(map[string]Output, len(outputType.Fields))
	for fieldName, fieldType := range outputType.Fields {
		sub.Fields[fieldName] = subOutput(output.Fields[fieldName], fieldType)
	}
	return
}

func (pi parInstance) Run(input Input, output Output) {
	for _, inner := range pi.Inners {
		innerInput := subInput(input, *inner.InputType())
		innerOutput := subOutput(output, *inner.OutputType())
		go inner.Run(innerInput, innerOutput)
	}
}
