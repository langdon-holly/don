package coms

import . "don/core"

type CompositeComChanSourceN struct {
	Units, Syntaxen, GenComs int
}

type CompositeComEntry struct {
	Com
	InputMap, OutputMap interface{}
}

// Inner chans must be mapped before outer chans
// One (1) chan per input
type CompositeCom struct {
	InputChanN  CompositeComChanSourceN
	OutputChanN CompositeComChanSourceN
	InnerChanN  CompositeComChanSourceN

	TheInputType, TheOutputType DType

	InputMap  interface{}
	OutputMap interface{}

	ComEntries []CompositeComEntry
}

func (com CompositeCom) InputType() DType {
	return com.TheInputType
}

func (com CompositeCom) OutputType() DType {
	return com.TheOutputType
}

type inputChanSource struct {
	Units    []<-chan Unit
	Syntaxen []<-chan Syntax
	GenComs  []<-chan GenCom
}

type outputChanSource struct {
	Units    []chan<- Unit
	Syntaxen []chan<- Syntax
	GenComs  []chan<- GenCom
}

func makeInputChanSource(n CompositeComChanSourceN) (ret inputChanSource) {
	ret.Units = make([]<-chan Unit, n.Units)
	ret.Syntaxen = make([]<-chan Syntax, n.Syntaxen)
	ret.GenComs = make([]<-chan GenCom, n.GenComs)
	return
}

func makeOutputChanSource(n CompositeComChanSourceN) (ret outputChanSource) {
	ret.Units = make([]chan<- Unit, n.Units)
	ret.Syntaxen = make([]chan<- Syntax, n.Syntaxen)
	ret.GenComs = make([]chan<- GenCom, n.GenComs)
	return
}

func putInputChans(dType DType, chanMap interface{}, input Input, chans inputChanSource) {
	switch dType.Tag {
	case UnitTypeTag:
		chans.Units[chanMap.(int)] = input.Unit
	case SyntaxTypeTag:
		chans.Syntaxen[chanMap.(int)] = input.Syntax
	case GenComTypeTag:
		chans.GenComs[chanMap.(int)] = input.GenCom
	case StructTypeTag:
		for fieldName, fieldType := range dType.Fields {
			putInputChans(fieldType, chanMap.(map[string]interface{})[fieldName], input.Struct[fieldName], chans)
		}
	}
}

func putOutputChans(dType DType, chanMap interface{}, output Output, chans outputChanSource) {
	switch dType.Tag {
	case UnitTypeTag:
		chans.Units[chanMap.(int)] = output.Unit
	case SyntaxTypeTag:
		chans.Syntaxen[chanMap.(int)] = output.Syntax
	case GenComTypeTag:
		chans.GenComs[chanMap.(int)] = output.GenCom
	case StructTypeTag:
		for fieldName, fieldType := range dType.Fields {
			putOutputChans(fieldType, chanMap.(map[string]interface{})[fieldName], output.Struct[fieldName], chans)
		}
	}
}

func getInput(dType DType, chanMap interface{}, chans inputChanSource) (ret Input) {
	switch dType.Tag {
	case UnitTypeTag:
		ret.Unit = chans.Units[chanMap.(int)]
	case SyntaxTypeTag:
		ret.Syntax = chans.Syntaxen[chanMap.(int)]
	case GenComTypeTag:
		ret.GenCom = chans.GenComs[chanMap.(int)]
	case StructTypeTag:
		ret.Struct = make(StructIn)
		for fieldName, fieldType := range dType.Fields {
			ret.Struct[fieldName] = getInput(fieldType, chanMap.(map[string]interface{})[fieldName], chans)
		}
	default:
		panic("Unreachable")
	}

	return
}

func getOutput(dType DType, chanMap interface{}, chans outputChanSource) (ret Output) {
	switch dType.Tag {
	case UnitTypeTag:
		ret.Unit = chans.Units[chanMap.(int)]
	case SyntaxTypeTag:
		ret.Syntax = chans.Syntaxen[chanMap.(int)]
	case GenComTypeTag:
		ret.GenCom = chans.GenComs[chanMap.(int)]
	case StructTypeTag:
		ret.Struct = make(StructOut)
		for fieldName, fieldType := range dType.Fields {
			ret.Struct[fieldName] = getOutput(fieldType, chanMap.(map[string]interface{})[fieldName], chans)
		}
	default:
		panic("Unreachable")
	}

	return
}

func (com CompositeCom) Run(input Input, output Output, quit <-chan struct{}) {
	inChans := makeInputChanSource(com.InputChanN)
	outChans := makeOutputChanSource(com.OutputChanN)

	putInputChans(com.TheInputType, com.InputMap, input, inChans)
	putOutputChans(com.TheOutputType, com.OutputMap, output, outChans)

	for i := 0; i < com.InnerChanN.Units; i++ {
		theChan := make(chan Unit, 1)
		inChans.Units[i] = (<-chan Unit)(theChan)
		outChans.Units[i] = chan<- Unit(theChan)
	}
	for i := 0; i < com.InnerChanN.Syntaxen; i++ {
		theChan := make(chan Syntax, 1)
		inChans.Syntaxen[i] = (<-chan Syntax)(theChan)
		outChans.Syntaxen[i] = chan<- Syntax(theChan)
	}
	for i := 0; i < com.InnerChanN.GenComs; i++ {
		theChan := make(chan GenCom, 1)
		inChans.GenComs[i] = (<-chan GenCom)(theChan)
		outChans.GenComs[i] = chan<- GenCom(theChan)
	}

	for _, comEntry := range com.ComEntries {
		input := getInput(comEntry.InputType(), comEntry.InputMap, inChans)
		output := getOutput(comEntry.OutputType(), comEntry.OutputMap, outChans)
		go comEntry.Run(input, output, quit)
	}
}

func MakeCompositeComMaps(map0, map1 *interface{}, chanN *CompositeComChanSourceN, theType DType) {
	switch theType.Tag {
	case UnitTypeTag:
		*map0 = chanN.Units
		*map1 = chanN.Units
		chanN.Units++
	case SyntaxTypeTag:
		*map0 = chanN.Syntaxen
		*map1 = chanN.Syntaxen
		chanN.Syntaxen++
	case GenComTypeTag:
		*map0 = chanN.GenComs
		*map1 = chanN.GenComs
		chanN.GenComs++
	case StructTypeTag:
		map0Val := make(map[string]interface{})
		*map0 = map0Val

		map1Val := make(map[string]interface{})
		*map1 = map1Val

		for fieldName, fieldType := range theType.Fields {
			var fieldMap0 interface{}
			var fieldMap1 interface{}

			MakeCompositeComMaps(&fieldMap0, &fieldMap1, chanN, fieldType)

			map0Val[fieldName] = fieldMap0
			map1Val[fieldName] = fieldMap1
		}
	}
}

type ReaderId struct {
	InternalP   bool
	InternalIdx int /* when InternalP */
}

type SignalReaderId struct {
	ReaderId
	FieldPath []string
}

type SignalReaderIdTree struct {
	ParentP        bool
	Children       map[string]SignalReaderIdTree /* when ParentP */
	SignalReaderId                               /* when !ParentP */
}

type GenCompositeEntry struct {
	GenCom
	OutputMap SignalReaderIdTree
}

type GenComposite struct {
	GenComs  []GenCompositeEntry
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

func (gc GenComposite) InferTypes(inputPType PartialType) (out ReaderInputTypes) {
	out.Internals = make([]PartialType, len(gc.GenComs))

	waiters := make(map[int]struct{}, len(gc.GenComs))
	for i := 0; i < len(gc.GenComs); i++ {
		waiters[i] = struct{}{}
	}

	sendTypeToReaders(inputPType, gc.InputMap, &out, waiters)
	for {
		waiter, ok := grabInt(waiters)
		if !ok {
			break
		}
		entry := gc.GenComs[waiter]
		sendTypeToReaders(entry.OutputType(out.Internals[waiter]), entry.OutputMap, &out, waiters)
	}

	return
}

func (gc GenComposite) OutputType(inputType PartialType) PartialType {
	return gc.InferTypes(inputType).External
}

type externalityTree struct {
	ParentP   bool
	Children  map[string]*externalityTree /* when ParentP */
	ExternalP bool                        /* when !ParentP */
}

func initExternalityTrees(externalities []externalityTree, externalInputMap SignalReaderIdTree) {
	if externalInputMap.ParentP {
		for _, subMap := range externalInputMap.Children {
			initExternalityTrees(externalities, subMap)
		}
	} else {
		initExternalityTree(&externalities[externalInputMap.SignalReaderId.ReaderId.InternalIdx], externalInputMap.SignalReaderId.FieldPath)
	}
}

func initExternalityTree(externality *externalityTree, fieldPath []string) {
	for {
		if externality.ExternalP {
			return
		}

		if len(fieldPath) == 0 {
			*externality = externalityTree{ExternalP: true}
			return
		}

		if !externality.ParentP {
			externality.ParentP = true
			externality.Children = make(map[string]*externalityTree, 1)
		}

		subExternality, ok := externality.Children[fieldPath[0]]
		if !ok {
			subExternality = new(externalityTree)
			externality.Children[fieldPath[0]] = subExternality
		}

		externality = subExternality
		fieldPath = fieldPath[1:]
	}
}

func makeInputMapInnards(inputMap *interface{}, inputChanN *CompositeComChanSourceN, externality externalityTree, inputType DType) {
	if externality.ParentP {
		inputMapVal := make(map[string]interface{}, len(inputType.Fields))
		*inputMap = inputMapVal
		for fieldName, fieldType := range inputType.Fields {
			var subExternality externalityTree
			subExternalityPointer := externality.Children[fieldName]
			if subExternalityPointer != nil {
				subExternality = *subExternalityPointer
			}

			var subMap interface{}
			makeInputMapInnards(&subMap, inputChanN, subExternality, fieldType)
			inputMapVal[fieldName] = subMap
		}
	} else if !externality.ExternalP {
		MakeCompositeComMaps(inputMap, new(interface{}), inputChanN, inputType)
	}
}

func makeInputMapExternals(inputMap *interface{}, inputChanN *CompositeComChanSourceN, externality externalityTree, inputType DType) {
	if externality.ParentP {
		inputMapVal := (*inputMap).(map[string]interface{})
		for fieldName, fieldType := range inputType.Fields {
			var subExternality externalityTree
			subExternalityPointer := externality.Children[fieldName]
			if subExternalityPointer != nil {
				subExternality = *subExternalityPointer
			}

			subMap := inputMapVal[fieldName]
			makeInputMapExternals(&subMap, inputChanN, subExternality, fieldType)
			inputMapVal[fieldName] = subMap
		}
	} else if externality.ExternalP {
		MakeCompositeComMaps(inputMap, new(interface{}), inputChanN, inputType)
	}
}

func makeOutputMap(genOutputMap SignalReaderIdTree, entries []CompositeComEntry, externalOutputMap interface{}) interface{} {
	if genOutputMap.ParentP {
		fields := make(map[string]interface{}, len(genOutputMap.Children))
		for fieldName, subMap := range genOutputMap.Children {
			fields[fieldName] = makeOutputMap(subMap, entries, externalOutputMap)
		}
		return fields
	} else {
		id := genOutputMap.SignalReaderId

		var inputMap interface{}
		if id.ReaderId.InternalP {
			inputMap = entries[id.ReaderId.InternalIdx].InputMap
		} else {
			inputMap = externalOutputMap
		}

		for _, fieldName := range id.FieldPath {
			inputMap = inputMap.(map[string]interface{})[fieldName]
		}
		return inputMap
	}
}

func (gc GenComposite) Com(inputType DType) Com {
	readerInputTypes := gc.InferTypes(PartializeType(inputType))

	outputType := HolizePartialType(readerInputTypes.External)

	comEntries := make([]CompositeComEntry, len(gc.GenComs))
	for i, entry := range gc.GenComs {
		comEntries[i].Com = entry.Com(HolizePartialType(readerInputTypes.Internals[i]))
	}

	// Make reader maps and chanNs

	externalityTrees := make([]externalityTree, len(gc.GenComs))
	initExternalityTrees(externalityTrees, gc.InputMap)

	var innerChanN CompositeComChanSourceN
	for i, entry := range comEntries {
		makeInputMapInnards(&comEntries[i].InputMap, &innerChanN, externalityTrees[i], entry.InputType())
	}

	inputChanN := innerChanN
	for i, entry := range comEntries {
		makeInputMapExternals(&comEntries[i].InputMap, &inputChanN, externalityTrees[i], entry.InputType())
	}

	outputChanN := innerChanN
	var outputMap interface{}
	MakeCompositeComMaps(&outputMap, new(interface{}), &outputChanN, outputType)

	// Make writer maps

	inputMap := makeOutputMap(gc.InputMap, comEntries, outputMap)
	for i, genCom := range gc.GenComs {
		comEntries[i].OutputMap = makeOutputMap(genCom.OutputMap, comEntries, outputMap)
	}

	// Return

	return CompositeCom{
		InputChanN:  inputChanN,
		OutputChanN: outputChanN,
		InnerChanN:  innerChanN,

		TheInputType:  inputType,
		TheOutputType: outputType,

		InputMap:  inputMap,
		OutputMap: outputMap,

		ComEntries: comEntries}
}
