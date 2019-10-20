package coms

import . "don/core"
import "don/extra"

type ReaderId struct {
	InternalP   bool
	InternalIdx int /* for InternalP */
}

type SignalReaderId struct {
	ReaderId
	FieldPath []string
}

type SignalReaderIdTree struct {
	ParentP        bool
	Children       map[string]SignalReaderIdTree /* for ParentP */
	SignalReaderId                               /* for !ParentP */
}

type CompositeComEntry struct {
	Com
	OutputMap SignalReaderIdTree
}

type CompositeCom struct {
	Coms     []CompositeComEntry
	InputMap SignalReaderIdTree
}

type ReaderInputTypes struct {
	Internals []PartialType
	External  PartialType
}

func sendTypeToReaders(pType PartialType, outputMap SignalReaderIdTree, readerInputTypes *ReaderInputTypes, waiters map[int]struct{}) {
	if outputMap.ParentP {
		if !pType.P {
			return
		}

		for fieldName, innerOutputMap := range outputMap.Children {
			sendTypeToReaders(pType.Fields[fieldName], innerOutputMap, readerInputTypes, waiters)
		}
	} else {
		readerInputType := PartialTypeAtPath(pType, outputMap.FieldPath)
		if outputMap.InternalP {
			idx := outputMap.InternalIdx
			oldInputType := readerInputTypes.Internals[idx]
			newInputType := MergePartialTypes(oldInputType, readerInputType)
			if !oldInputType.Equal(newInputType) {
				readerInputTypes.Internals[idx] = newInputType
				waiters[idx] = struct{}{}
			}
		} else {
			readerInputTypes.External = MergePartialTypes(readerInputTypes.External, readerInputType)
		}
	}
}

func grabInt(ints map[int]struct{}) (int, bool) {
	for grabbed, _ := range ints {
		delete(ints, grabbed)
		return grabbed, true
	}
	return 0, false
}

func (gc CompositeCom) InferTypes(inputPType PartialType) (out ReaderInputTypes) {
	out.Internals = make([]PartialType, len(gc.Coms))

	waiters := make(map[int]struct{}, len(gc.Coms))
	for i := 0; i < len(gc.Coms); i++ {
		waiters[i] = struct{}{}
	}

	sendTypeToReaders(inputPType, gc.InputMap, &out, waiters)
	for {
		waiter, ok := grabInt(waiters)
		if !ok {
			break
		}
		entry := gc.Coms[waiter]
		sendTypeToReaders(entry.OutputType(out.Internals[waiter]), entry.OutputMap, &out, waiters)
	}

	return
}

func (gc CompositeCom) OutputType(inputType PartialType) PartialType {
	return gc.InferTypes(inputType).External
}

func putExternalInput(inputMap SignalReaderIdTree, input Input, inputIs []Input /* mutated */) {
	if inputMap.ParentP {
		for fieldName, innerInputMap := range inputMap.Children {
			putExternalInput(innerInputMap, input.Struct[fieldName], inputIs)
		}
	} else {
		if len(inputMap.FieldPath) == 0 {
			inputIs[inputMap.InternalIdx] = input
		} else {
			parentStruct := inputIs[inputMap.InternalIdx].Struct
			for _, fieldName := range inputMap.FieldPath[:len(inputMap.FieldPath)-1] {
				parentStruct = parentStruct[fieldName].Struct
			}
			parentStruct[inputMap.FieldPath[len(inputMap.FieldPath)]] = input
		}
	}
}

func getInternalOutput(outputMap SignalReaderIdTree, inputOs []Output, externalOutput Output) (ret Output) {
	if outputMap.ParentP {
		ret.Struct = make(map[string]Output, len(outputMap.Children))
		for fieldName, innerOutputMap := range outputMap.Children {
			ret.Struct[fieldName] = getInternalOutput(innerOutputMap, inputOs, externalOutput)
		}
	} else {
		if outputMap.InternalP {
			ret = inputOs[outputMap.InternalIdx]
		} else {
			ret = externalOutput
		}

		for _, fieldName := range outputMap.FieldPath {
			ret = ret.Struct[fieldName]
		}
	}

	return
}

func (gc CompositeCom) Run(inputType DType, input Input, output Output, quit <-chan struct{}) {
	inputTypes := gc.InferTypes(PartializeType(inputType))

	inputIs := make([]Input, len(gc.Coms))
	inputOs := make([]Output, len(gc.Coms))
	for i, inputType := range inputTypes.Internals {
		subInputs, subOutput := extra.MakeIOChans(HolizePartialType(inputType), 1)
		inputIs[i] = subInputs[0]
		inputOs[i] = subOutput
	}
	putExternalInput(gc.InputMap, input, inputIs)

	outputOs := make([]Output, len(gc.Coms))
	for i, entry := range gc.Coms {
		outputOs[i] = getInternalOutput(entry.OutputMap, inputOs, output)
	}

	for i, com := range gc.Coms {
		go com.Run(HolizePartialType(inputTypes.Internals[i]), inputIs[i], outputOs[i], quit)
	}
}
