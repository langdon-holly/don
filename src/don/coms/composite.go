package coms

import . "don/core"

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
	Internals []DType
	External  DType
}

func sendTypeToReaders(pType DType, outputMap SignalReaderIdTree, readerInputTypes *ReaderInputTypes, waiters map[int]struct{}) {
	if outputMap.ParentP {
		if !pType.P {
			return
		}

		for fieldName, innerOutputMap := range outputMap.Children {
			sendTypeToReaders(pType.Fields[fieldName], innerOutputMap, readerInputTypes, waiters)
		}
	} else {
		readerInputType := TypeAtPath(pType, outputMap.FieldPath)
		if outputMap.InternalP {
			idx := outputMap.InternalIdx
			oldInputType := readerInputTypes.Internals[idx]
			newInputType := MergeTypes(oldInputType, readerInputType)
			if !oldInputType.Equal(newInputType) {
				readerInputTypes.Internals[idx] = newInputType
				waiters[idx] = struct{}{}
			}
		} else {
			readerInputTypes.External = MergeTypes(readerInputTypes.External, readerInputType)
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

func (gc CompositeCom) InferTypes(inputPType DType) (out ReaderInputTypes) {
	out.Internals = make([]DType, len(gc.Coms))

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

func (gc CompositeCom) OutputType(inputType DType) DType {
	return gc.InferTypes(inputType).External
}

func putExternalInput(inputMap SignalReaderIdTree, inputGetter InputGetter, inputIGetters []InputGetter /* mutated */) {
	if inputMap.ParentP {
		for fieldName, innerInputMap := range inputMap.Children {
			putExternalInput(innerInputMap, inputGetter.Struct[fieldName], inputIGetters)
		}
	} else {
		if len(inputMap.FieldPath) == 0 {
			inputIGetters[inputMap.InternalIdx] = inputGetter
		} else {
			parentStruct := inputIGetters[inputMap.InternalIdx].Struct
			for _, fieldName := range inputMap.FieldPath[:len(inputMap.FieldPath)-1] {
				parentStruct = parentStruct[fieldName].Struct
			}
			parentStruct[inputMap.FieldPath[len(inputMap.FieldPath)]] = inputGetter
		}
	}
}

func getInternalOutput(outputMap SignalReaderIdTree, inputOGetters []OutputGetter, externalOutputGetter OutputGetter) (ret OutputGetter) {
	if outputMap.ParentP {
		ret.Struct = make(map[string]OutputGetter, len(outputMap.Children))
		for fieldName, innerOutputMap := range outputMap.Children {
			ret.Struct[fieldName] = getInternalOutput(innerOutputMap, inputOGetters, externalOutputGetter)
		}
	} else {
		if outputMap.InternalP {
			ret = inputOGetters[outputMap.InternalIdx]
		} else {
			ret = externalOutputGetter
		}

		for _, fieldName := range outputMap.FieldPath {
			ret = ret.Struct[fieldName]
		}
	}

	return
}

func (gc CompositeCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	inputTypes := gc.InferTypes(inputType)

	inputIGetters := make([]InputGetter, len(gc.Coms))
	inputOGetters := make([]OutputGetter, len(gc.Coms))
	for i, inputType := range inputTypes.Internals {
		inputIGetters[i], inputOGetters[i] = MakeIO(inputType)
	}
	putExternalInput(gc.InputMap, inputGetter, inputIGetters)

	outputOGetters := make([]OutputGetter, len(gc.Coms))
	for i, entry := range gc.Coms {
		outputOGetters[i] = getInternalOutput(entry.OutputMap, inputOGetters, outputGetter)
	}

	for i, com := range gc.Coms {
		go com.Run(inputTypes.Internals[i], inputIGetters[i], outputOGetters[i], quit)
	}
}
