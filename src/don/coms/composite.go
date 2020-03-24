package coms

import . "don/core"

type SignalMap struct {
	ParentP bool

	// for ParentP
	Children map[string]SignalMap

	// for !ParentP
	ExternalP   bool
	InternalIdx int /* for !ExternalP */
	FieldPath   []string
}

type CompositeComEntry struct {
	Com
	OutputMap SignalMap
}

type CompositeCom struct {
	Coms     []CompositeComEntry
	InputMap SignalMap
}

type ReaderInputTypes struct {
	Internals []DType
	External  DType
}

func sendTypeToReaders(theType DType, outputMap SignalMap, readerInputTypes *ReaderInputTypes, waiters map[int]struct{}) {
	if outputMap.ParentP {
		switch theType.Lvl {
		case UnknownTypeLvl:
		case NormalTypeLvl:
			for fieldName, innerOutputMap := range outputMap.Children {
				sendTypeToReaders(theType.Fields[fieldName], innerOutputMap, readerInputTypes, waiters)
			}
		case ImpossibleTypeLvl:
			for _, innerOutputMap := range outputMap.Children {
				sendTypeToReaders(ImpossibleType, innerOutputMap, readerInputTypes, waiters)
			}
		}
	} else {
		readerInputType := TypeAtPath(theType, outputMap.FieldPath)
		if outputMap.ExternalP {
			readerInputTypes.External = MergeTypes(readerInputTypes.External, readerInputType)
		} else {
			idx := outputMap.InternalIdx
			oldInputType := readerInputTypes.Internals[idx]
			newInputType := MergeTypes(oldInputType, readerInputType)
			if !oldInputType.Equal(newInputType) {
				readerInputTypes.Internals[idx] = newInputType
				waiters[idx] = struct{}{}
			}
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

func (cc CompositeCom) InferTypes(inputType DType) (out ReaderInputTypes) {
	out.Internals = make([]DType, len(cc.Coms))

	waiters := make(map[int]struct{}, len(cc.Coms))
	for i := 0; i < len(cc.Coms); i++ {
		waiters[i] = struct{}{}
	}

	sendTypeToReaders(inputType, cc.InputMap, &out, waiters)
	for {
		waiter, ok := grabInt(waiters)
		if !ok {
			break
		}
		entry := cc.Coms[waiter]
		sendTypeToReaders(entry.OutputType(out.Internals[waiter]), entry.OutputMap, &out, waiters)
	}

	return
}

func (cc CompositeCom) OutputType(inputType DType) DType {
	return cc.InferTypes(inputType).External
}

func putExternalInput(inputMap SignalMap, inputGetter InputGetter, inputIGetters []InputGetter /* mutated */) {
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

func getInternalOutput(outputMap SignalMap, inputOGetters []OutputGetter, externalOutputGetter OutputGetter) (ret OutputGetter) {
	if outputMap.ParentP {
		ret.Struct = make(map[string]OutputGetter, len(outputMap.Children))
		for fieldName, innerOutputMap := range outputMap.Children {
			ret.Struct[fieldName] = getInternalOutput(innerOutputMap, inputOGetters, externalOutputGetter)
		}
	} else {
		if outputMap.ExternalP {
			ret = externalOutputGetter
		} else {
			ret = inputOGetters[outputMap.InternalIdx]
		}

		for _, fieldName := range outputMap.FieldPath {
			ret = ret.Struct[fieldName]
		}
	}

	return
}

func (cc CompositeCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	inputTypes := cc.InferTypes(inputType)

	inputIGetters := make([]InputGetter, len(cc.Coms))
	inputOGetters := make([]OutputGetter, len(cc.Coms))
	for i, inputType := range inputTypes.Internals {
		inputIGetters[i], inputOGetters[i] = MakeIO(inputType)
	}
	putExternalInput(cc.InputMap, inputGetter, inputIGetters)

	outputOGetters := make([]OutputGetter, len(cc.Coms))
	for i, entry := range cc.Coms {
		outputOGetters[i] = getInternalOutput(entry.OutputMap, inputOGetters, outputGetter)
	}

	for i, com := range cc.Coms {
		go com.Run(inputTypes.Internals[i], inputIGetters[i], outputOGetters[i], quit)
	}
}
