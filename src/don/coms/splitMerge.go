package coms

import "strconv"

import . "don/core"

type SplitMergeCom []Com

const (
	splitMergeComSubComIdInnerTag = iota
	splitMergeComSubComIdSplitTag
	splitMergeComSubComIdMergeTag
)

type splitMergeComSubComId struct {
	Tag int
	Idx int /* for Tag == splitMergeComSubComIdInnerTag */
}

func (smc SplitMergeCom) types(inputType, outputType *DType) (done bool, splitOutputType, mergeInputType DType) {
	indexStrings := make([]string, len(smc))
	splitOutputType, mergeInputType = MakeNStructType(len(smc)), MakeNStructType(len(smc))

	innerDones := make([]bool, len(smc))
	splitDone := false
	mergeDone := false

	toType := make(map[splitMergeComSubComId]struct{}, len(smc)+2)
	toType[splitMergeComSubComId{Tag: splitMergeComSubComIdSplitTag}] = struct{}{}
	toType[splitMergeComSubComId{Tag: splitMergeComSubComIdMergeTag}] = struct{}{}
	for i := range smc {
		indexStrings[i] = strconv.FormatInt(int64(i), 10)
		splitOutputType.Fields[indexStrings[i]] = UnknownType
		mergeInputType.Fields[indexStrings[i]] = UnknownType
		toType[splitMergeComSubComId{Idx: i}] = struct{}{}
	}
	for len(toType) > 0 {
		var typeNext splitMergeComSubComId
		for typeNext = range toType {
			delete(toType, typeNext)
			break
		}
		switch typeNext.Tag {
		case splitMergeComSubComIdInnerTag:
			innerInputType := splitOutputType.Fields[indexStrings[typeNext.Idx]]
			innerOutputType := mergeInputType.Fields[indexStrings[typeNext.Idx]]
			innerInputTypeBefore := innerInputType
			innerOutputTypeBefore := innerOutputType

			innerDones[typeNext.Idx] = smc[typeNext.Idx].Types(&innerInputType, &innerOutputType)
			splitOutputType.Fields[indexStrings[typeNext.Idx]] = innerInputType
			mergeInputType.Fields[indexStrings[typeNext.Idx]] = innerOutputType

			if !innerInputTypeBefore.Equal(innerInputType) {
				toType[splitMergeComSubComId{Tag: splitMergeComSubComIdSplitTag}] = struct{}{}
			}
			if !innerOutputTypeBefore.Equal(innerOutputType) {
				toType[splitMergeComSubComId{Tag: splitMergeComSubComIdMergeTag}] = struct{}{}
			}
		case splitMergeComSubComIdSplitTag:
			splitOutputTypeBefore := splitOutputType
			splitDone = SplitCom{}.Types(inputType, &splitOutputType)
			for i := range smc {
				if !splitOutputTypeBefore.Fields[indexStrings[i]].Equal(splitOutputType.Fields[indexStrings[i]]) {
					toType[splitMergeComSubComId{Idx: i}] = struct{}{}
				}
			}
		case splitMergeComSubComIdMergeTag:
			mergeInputTypeBefore := mergeInputType
			mergeDone = MergeCom{}.Types(&mergeInputType, outputType)
			for i := range smc {
				if !mergeInputTypeBefore.Fields[indexStrings[i]].Equal(mergeInputType.Fields[indexStrings[i]]) {
					toType[splitMergeComSubComId{Idx: i}] = struct{}{}
				}
			}
		}
	}

	done = splitDone && mergeDone
	for _, innerDone := range innerDones {
		done = done && innerDone
	}
	return
}

func (smc SplitMergeCom) Types(inputType, outputType *DType) (done bool) {
	done, _, _ = smc.types(inputType, outputType)
	return
}

func (smc SplitMergeCom) Run(inputType, outputType DType, input Input, output Output) {
	_, splitOutputType, mergeInputType := smc.types(&inputType, &outputType)
	splitOutput := Output{Fields: make(map[string]Output, len(smc))}
	mergeInput := Input{Fields: make(map[string]Input, len(smc))}
	for i := range smc {
		fieldName := strconv.FormatInt(int64(i), 10)
		innerInputInput, innerInputOutput := MakeIO(splitOutputType.Fields[fieldName])
		innerOutputInput, innerOutputOutput := MakeIO(mergeInputType.Fields[fieldName])
		splitOutput.Fields[fieldName] = innerInputOutput
		mergeInput.Fields[fieldName] = innerOutputInput
		go smc[i].Run(splitOutputType.Fields[fieldName], mergeInputType.Fields[fieldName], innerInputInput, innerOutputOutput)
	}
	go SplitCom{}.Run(inputType, splitOutputType, input, splitOutput)
	go MergeCom{}.Run(mergeInputType, outputType, mergeInput, output)
}
