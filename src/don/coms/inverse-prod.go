package coms

import "strconv"

import . "don/core"

func InverseProd() Com { return &InverseProdCom{Prod: Prod()} }

type InverseProdCom struct{ Prod Com }

func (ipc InverseProdCom) InputType() *DType  { return ipc.Prod.OutputType() }
func (ipc InverseProdCom) OutputType() *DType { return ipc.Prod.InputType() }

func (ipc InverseProdCom) Types() Com {
	ipc.Prod = ipc.Prod.Types()
	if _, nullp := ipc.Prod.(NullCom); nullp {
		return Null
	} else {
		return ipc
	}
}

func (ipc InverseProdCom) Underdefined() Error {
	return ipc.Prod.Underdefined().Context("in inverse prod")
}

func (ipc InverseProdCom) Copy() Com { ipc.Prod = ipc.Prod.Copy(); return ipc }

func (ipc InverseProdCom) Invert() Com { return ipc.Prod }

func runForks(
	mergeInput Input, /* mutated */
	mergeInputFieldName string,
	outputType DType,
	outputFieldIdx int,
	outputFieldType DType, /* for outputFieldIdx < len(outputType.Fields) */
	forkInput Input,
	forkOutputType DType) {
	if outputFieldIdx == len(outputType.Fields) {
		fork := Fork()
		fork.InputType().Meets(UnitType)
		fork.OutputType().Meets(forkOutputType)
		if underdefined := fork.Underdefined(); underdefined != nil {
			panic("Unreachable underdefined:\n" + underdefined.String())
		}
		//theseMergeInput, forkOutput := MakeIO(forkOutputType)
		//go fork.Run(forkInput, forkOutput)

	} else {
		for fieldName, subForkInput := range forkInput.Fields {
			subMergeInputFieldName := mergeInputFieldName + "," + fieldName
			subOutputFieldIdx := outputFieldIdx
			var subOutputFieldType DType
			if outputFieldType.NoUnit {
				for _, subOutputFieldType = range outputFieldType.Fields {
					break
				}
			} else {
				subOutputFieldIdx++
				if subOutputFieldIdx < len(outputType.Fields) {
					subOutputFieldType = outputType.Fields[strconv.Itoa(subOutputFieldIdx)]
				}
			}
			runForks(
				mergeInput,
				subMergeInputFieldName,
				outputType,
				subOutputFieldIdx,
				subOutputFieldType,
				subForkInput,
				forkOutputType)
		}
	}
}

// leaf may be shared
func mergeDeep(inputType DType, leaf Com) Com {
	if inputType.NoUnit {
		subs := make([]Com, len(inputType.Fields))
		i := 0
		for fieldName, fieldType := range inputType.Fields {
			subs[i] = Pipe([]Com{Select(fieldName), mergeDeep(fieldType, leaf)})
			i++
		}
		return Pipe([]Com{Scatter(), Par(subs), Merge()})
	} else {
		return leaf.Copy()
	}
}

// leaf may be shared
func mapDeep(inputType DType, leaf Com) Com {
	if inputType.NoUnit {
		subs := make([]Com, len(inputType.Fields))
		i := 0
		for fieldName, fieldType := range inputType.Fields {
			subs[i] = Pipe([]Com{
				Select(fieldName),
				mapDeep(fieldType, leaf),
				Deselect(fieldName)})
			i++
		}
		return Pipe([]Com{Scatter(), Par(subs), Gather()})
	} else {
		return leaf.Copy()
	}
}

func (ipc InverseProdCom) Run(input Input, output Output) {
	outputType := *ipc.Prod.InputType()
	if len(outputType.Fields) == 0 {
		return
	}
	inputType := *ipc.Prod.OutputType()

	merges := make([]Com, len(outputType.Fields))
	for i := 0; i < len(outputType.Fields); i++ {
		merges[i] = I(UnitType)
		for j := len(outputType.Fields) - 1; j > i; j-- {
			merges[i] = mergeDeep(outputType.Fields[strconv.Itoa(j)], merges[i])
		}
		merges[i] = mapDeep(outputType.Fields[strconv.Itoa(i)], merges[i])
		for j := i - 1; j >= 0; j-- {
			merges[i] = mergeDeep(outputType.Fields[strconv.Itoa(j)], merges[i])
		}
	}
	com := Pipe([]Com{Fork(), Par(merges)})
	com.InputType().Meets(inputType)
	com.OutputType().Meets(outputType)
	com = com.Types()
	if underdefined := com.Underdefined(); underdefined != nil {
		panic("Unreachable underdefined:\n" + underdefined.String())
	} else if !inputType.LTE(*com.InputType()) {
		panic("Unreachable")
	} else if !outputType.LTE(*com.OutputType()) {
		panic("Unreachable:\n" + com.OutputType().String() + "\nnot\n" + outputType.String())
	}
	com.Run(input, output)
}
