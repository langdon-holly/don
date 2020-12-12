package coms

import "strconv"

import . "don/core"

type InverseProdCom struct{}

func (InverseProdCom) Instantiate() ComInstance {
	return inverseProdInstance{Prod: ProdCom{}.Instantiate()}
}

func (InverseProdCom) Inverse() Com { return ProdCom{} }

type inverseProdInstance struct{ Prod ComInstance }

func (ipi inverseProdInstance) InputType() *DType  { return ipi.Prod.OutputType() }
func (ipi inverseProdInstance) OutputType() *DType { return ipi.Prod.InputType() }

func (ipi inverseProdInstance) Types() { ipi.Prod.Types() }

func (ipi inverseProdInstance) Underdefined() Error {
	return ipi.Prod.Underdefined().Context("in inverse prod")
}

func runForks(
	mergeInput Input, /* mutated */
	mergeInputFieldName string,
	outputType DType,
	outputFieldIdx int,
	outputFieldType DType, /* for outputFieldIdx < len(outputType.Fields) */
	forkInput Input,
	forkOutputType DType) {
	if outputFieldIdx == len(outputType.Fields) {
		fork := ForkCom{}.Instantiate()
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

func mergeDeep(inputType DType, leaf Com) Com {
	if inputType.NoUnit {
		subs := make([]Com, len(inputType.Fields))
		i := 0
		for fieldName, fieldType := range inputType.Fields {
			subs[i] = PipeCom([]Com{SelectCom(fieldName), mergeDeep(fieldType, leaf)})
			i++
		}
		return PipeCom([]Com{ScatterCom{}, ParCom(subs), MergeCom{}})
	} else {
		return leaf
	}
}
func mapDeep(inputType DType, leaf Com) Com {
	if inputType.NoUnit {
		subs := make([]Com, len(inputType.Fields))
		i := 0
		for fieldName, fieldType := range inputType.Fields {
			subs[i] = PipeCom([]Com{
				SelectCom(fieldName),
				mapDeep(fieldType, leaf),
				DeselectCom(fieldName)})
			i++
		}
		return PipeCom([]Com{ScatterCom{}, ParCom(subs), GatherCom{}})
	} else {
		return leaf
	}
}

func (ipi inverseProdInstance) Run(input Input, output Output) {
	outputType := *ipi.Prod.InputType()
	if len(outputType.Fields) == 0 {
		return
	}
	inputType := *ipi.Prod.OutputType()

	merges := make([]Com, len(outputType.Fields))
	for i := 0; i < len(outputType.Fields); i++ {
		merges[i] = ICom(UnitType)
		for j := len(outputType.Fields) - 1; j > i; j-- {
			merges[i] = mergeDeep(outputType.Fields[strconv.Itoa(j)], merges[i])
		}
		merges[i] = mapDeep(outputType.Fields[strconv.Itoa(i)], merges[i])
		for j := i - 1; j >= 0; j-- {
			merges[i] = mergeDeep(outputType.Fields[strconv.Itoa(j)], merges[i])
		}
	}
	comI := PipeCom([]Com{ForkCom{}, ParCom(merges)}).Instantiate()
	comI.InputType().Meets(inputType)
	comI.OutputType().Meets(outputType)
	comI.Types()
	if underdefined := comI.Underdefined(); underdefined != nil {
		panic("Unreachable underdefined:\n" + underdefined.String())
	} else if !inputType.LTE(*comI.InputType()) {
		panic("Unreachable")
	} else if !outputType.LTE(*comI.OutputType()) {
		panic("Unreachable")
	}
	comI.Run(input, output)
}
