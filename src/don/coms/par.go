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

func (pc ParCom) types(inputType, outputType *DType) (
	underdefined Error, innerInputTypes, innerOutputTypes []DType) {
	var fanOutUnderdefined, fanInUnderdefined Error
	innerUnderdefineds := make([]Error, len(pc))
	innerInputTypes, innerOutputTypes = make([]DType, len(pc)), make([]DType, len(pc))

	toType := make(map[parComSubComId]struct{}, len(pc)+2)
	toType[parComSubComId{Tag: parComSubComIdFanOutTag}] = struct{}{}
	toType[parComSubComId{Tag: parComSubComIdFanInTag}] = struct{}{}
	for i := range pc {
		toType[parComSubComId{Idx: i}] = struct{}{}
	}
	for len(toType) > 0 {
		var typeNext parComSubComId
		for typeNext = range toType {
			delete(toType, typeNext)
			break
		}
		switch typeNext.Tag {
		case parComSubComIdInnerTag:
			innerInputTypeBefore := innerInputTypes[typeNext.Idx]
			innerOutputTypeBefore := innerOutputTypes[typeNext.Idx]
			innerUnderdefineds[typeNext.Idx] = pc[typeNext.Idx].Types(
				&innerInputTypes[typeNext.Idx], &innerOutputTypes[typeNext.Idx])
			if !innerInputTypeBefore.Equal(innerInputTypes[typeNext.Idx]) {
				toType[parComSubComId{Tag: parComSubComIdFanOutTag}] = struct{}{}
			}
			if !innerOutputTypeBefore.Equal(innerOutputTypes[typeNext.Idx]) {
				toType[parComSubComId{Tag: parComSubComIdFanInTag}] = struct{}{}
			}
		case parComSubComIdFanOutTag:
			innerInputTypesBefore := append([]DType{}, innerInputTypes...)
			fanOutUnderdefined = FanLinearTypes(innerInputTypes, inputType)
			for i := range pc {
				if !innerInputTypesBefore[i].Equal(innerInputTypes[i]) {
					toType[parComSubComId{Idx: i}] = struct{}{}
				}
			}
		case parComSubComIdFanInTag:
			innerOutputTypesBefore := append([]DType{}, innerOutputTypes...)
			fanInUnderdefined = FanLinearTypes(innerOutputTypes, outputType)
			for i := range pc {
				if !innerOutputTypesBefore[i].Equal(innerOutputTypes[i]) {
					toType[parComSubComId{Idx: i}] = struct{}{}
				}
			}
		}
	}

	underdefined.Ors(fanOutUnderdefined.Context("in par fan-out")).Ors(fanInUnderdefined.Context("in par fan-in"))
	for i, innerUnderdefined := range innerUnderdefineds {
		underdefined.Ors(innerUnderdefined.Context("in " + strconv.Itoa(i) + "'th computer in par"))
	}
	return
}

func (pc ParCom) Types(inputType, outputType *DType) (underdefined Error) {
	underdefined, _, _ = pc.types(inputType, outputType)
	return
}

// inputType.Positive
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

// outputType.Positive
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

func (pc ParCom) Run(inputType, outputType DType, input Input, output Output) {
	_, innerInputTypes, innerOutputTypes := pc.types(&inputType, &outputType)
	for i, inner := range pc {
		innerInput := subInput(input, innerInputTypes[i])
		innerOutput := subOutput(output, innerOutputTypes[i])
		go inner.Run(innerInputTypes[i], innerOutputTypes[i], innerInput, innerOutput)
	}
}
