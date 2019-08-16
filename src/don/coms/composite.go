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
	ret.GenComs = make([]<-chan GenCom, n.GenComs)
	return
}

func makeOutputChanSource(n CompositeComChanSourceN) (ret outputChanSource) {
	ret.Units = make([]chan<- Unit, n.Units)
	ret.GenComs = make([]chan<- GenCom, n.GenComs)
	return
}

func putInputChans(dType DType, chanMap interface{}, input interface{}, chans inputChanSource) {
	switch dType.Tag {
	case UnitTypeTag:
		chans.Units[chanMap.(int)] = input.(<-chan Unit)
	case SyntaxTypeTag:
		chans.Syntaxen[chanMap.(int)] = input.(<-chan Syntax)
	case GenComTypeTag:
		chans.GenComs[chanMap.(int)] = input.(<-chan GenCom)
	case StructTypeTag:
		for fieldName, fieldType := range dType.Extra.(map[string]DType) {
			putInputChans(fieldType, chanMap.(Struct)[fieldName], input.(Struct)[fieldName], chans)
		}
	}
}

func putOutputChans(dType DType, chanMap interface{}, input interface{}, chans outputChanSource) {
	switch dType.Tag {
	case UnitTypeTag:
		chans.Units[chanMap.(int)] = input.(chan<- Unit)
	case SyntaxTypeTag:
		chans.Syntaxen[chanMap.(int)] = input.(chan<- Syntax)
	case GenComTypeTag:
		chans.GenComs[chanMap.(int)] = input.(chan<- GenCom)
	case StructTypeTag:
		for fieldName, fieldType := range dType.Extra.(map[string]DType) {
			putOutputChans(fieldType, chanMap.(Struct)[fieldName], input.(Struct)[fieldName], chans)
		}
	}
}

func getInput(dType DType, chanMap interface{}, chans inputChanSource) interface{} {
	switch dType.Tag {
	case UnitTypeTag:
		return chans.Units[chanMap.(int)]
	case SyntaxTypeTag:
		return chans.Syntaxen[chanMap.(int)]
	case GenComTypeTag:
		return chans.GenComs[chanMap.(int)]
	case StructTypeTag:
		input := make(Struct)
		for fieldName, fieldType := range dType.Extra.(map[string]DType) {
			input[fieldName] = getInput(fieldType, chanMap.(Struct)[fieldName], chans)
		}
		return input
	default:
		panic("Unreachable")
	}
}

func getOutput(dType DType, chanMap interface{}, chans outputChanSource) interface{} {
	switch dType.Tag {
	case UnitTypeTag:
		return chans.Units[chanMap.(int)]
	case SyntaxTypeTag:
		return chans.Syntaxen[chanMap.(int)]
	case GenComTypeTag:
		return chans.GenComs[chanMap.(int)]
	case StructTypeTag:
		output := make(Struct)
		for fieldName, fieldType := range dType.Extra.(map[string]DType) {
			output[fieldName] = getOutput(fieldType, chanMap.(Struct)[fieldName], chans)
		}
		return output
	default:
		panic("Unreachable")
	}
}

func (com CompositeCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	inChans := makeInputChanSource(com.InputChanN)
	outChans := makeOutputChanSource(com.OutputChanN)

	putInputChans(com.TheInputType, com.InputMap, input, inChans)
	putOutputChans(com.TheOutputType, com.OutputMap, output, outChans)

	for i := 0; i < com.InnerChanN.Units; i++ {
		theChan := make(chan Unit, 1)
		inChans.Units[i] = (<-chan Unit)(theChan)
		outChans.Units[i] = chan<- Unit(theChan)
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
