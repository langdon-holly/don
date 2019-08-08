package main

import "fmt"

type DTypeTag int

const (
	UnitTypeTag = DTypeTag(iota)
	DTypeTypeTag
	ComTypeTag
	StructTypeTag
)

type Unit struct{}

type DType struct {
	Tag   DTypeTag
	Extra interface{}
}

var (
	UnitType  = DType{UnitTypeTag, nil}
	DTypeType = DType{DTypeTypeTag, nil}
)

func makeStructType(fields map[string]DType) DType {
	return DType{StructTypeTag, fields}
}

func makeComType(inputType, outputType DType) DType {
	return DType{ComTypeTag, [2]DType{inputType, outputType}}
}

var BoolTypeFields map[string]DType = make(map[string]DType, 2)
var BoolType DType = makeStructType(BoolTypeFields)

func makeIOChans(theType DType) (input, output interface{}) {
	switch theType.Tag {
	case UnitTypeTag:
		theChan := make(chan Unit, 1)
		input = (<-chan Unit)(theChan)
		output = chan<- Unit(theChan)
	case DTypeTypeTag:
		theChan := make(chan DType, 1)
		input = (<-chan DType)(theChan)
		output = chan<- DType(theChan)
	case ComTypeTag:
		theChan := make(chan Com, 1)
		input = (<-chan Com)(theChan)
		output = chan<- Com(theChan)
	case StructTypeTag:
		inputMap := make(map[string]interface{})
		outputMap := make(map[string]interface{})
		input = inputMap
		output = outputMap
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			inputMap[fieldName], outputMap[fieldName] = makeIOChans(fieldType)
		}
	}
	return
}

type ComTag int

const (
	AndComTag = ComTag(iota)
	MergeComTag
	SplitComTag
	ChooseComTag
	ConstComTag
	IComTag
	SelectComTag
	UnselectComTag
	CompositeComTag
)

type Com struct {
	Tag   ComTag
	Extra interface{}
}

type ConstVal struct {
	P   bool
	Val interface{}
}

type ConstComExtra struct {
	Type DType
	Val  interface{}
}

type SelectExtra struct {
	Fields    map[string]DType
	FieldName string
}

type UnselectExtra struct {
	FieldName string
	FieldType DType
}

type CompositeComChanSourceN struct {
	Units, Types, Coms int
}

type CompositeComEntry struct {
	Com
	InputMap, OutputMap interface{}
}

// Inner chans must be mapped before outer chans
// One (1) chan per input
type CompositeComExtra struct {
	InputChanN  CompositeComChanSourceN
	OutputChanN CompositeComChanSourceN
	InnerChanN  CompositeComChanSourceN

	InputType, OutputType DType

	InputMap  interface{}
	OutputMap interface{}

	ComEntries []CompositeComEntry
}

var andComInputTypeFields map[string]DType = make(map[string]DType, 2)
var andComInputType DType = makeStructType(andComInputTypeFields)

var chooseComInputTypeFields map[string]DType = make(map[string]DType, 3)
var chooseComInputType DType = makeStructType(chooseComInputTypeFields)

var chooseComOutputTypeFields map[string]DType = make(map[string]DType, 2)
var chooseComOutputType DType = makeStructType(chooseComOutputTypeFields)

var (
	AndCom    = Com{AndComTag, nil}
	ChooseCom = Com{ChooseComTag, nil}
)

func (com Com) InputType() DType {
	switch com.Tag {
	case AndComTag:
		return andComInputType
	case MergeComTag:
		theType := com.Extra.(DType)
		inputTypeExtra := make(map[string]DType, 2)
		inputTypeExtra["a"] = theType
		inputTypeExtra["b"] = theType
		return DType{StructTypeTag, inputTypeExtra}
	case SplitComTag:
		return com.Extra.(DType)
	case ChooseComTag:
		return chooseComInputType
	case ConstComTag:
		return UnitType
	case IComTag:
		return com.Extra.(DType)
	case SelectComTag:
		return makeStructType(com.Extra.(SelectExtra).Fields)
	case UnselectComTag:
		return com.Extra.(UnselectExtra).FieldType
	case CompositeComTag:
		return com.Extra.(CompositeComExtra).InputType
	default:
		panic("Unreachable")
	}
}

func (com Com) OutputType() DType {
	switch com.Tag {
	case AndComTag:
		return BoolType
	case MergeComTag:
		return com.Extra.(DType)
	case SplitComTag:
		theType := com.Extra.(DType)
		outputTypeExtra := make(map[string]DType, 2)
		outputTypeExtra["a"] = theType
		outputTypeExtra["b"] = theType
		return DType{StructTypeTag, outputTypeExtra}
	case ChooseComTag:
		return chooseComOutputType
	case ConstComTag:
		return com.Extra.(ConstComExtra).Type
	case IComTag:
		return com.Extra.(DType)
	case SelectComTag:
		extra := com.Extra.(SelectExtra)
		return extra.Fields[extra.FieldName]
	case UnselectComTag:
		extra := com.Extra.(UnselectExtra)
		fields := make(map[string]DType, 1)
		fields[extra.FieldName] = extra.FieldType
		return makeStructType(fields)
	case CompositeComTag:
		return com.Extra.(CompositeComExtra).InputType
	default:
		panic("Unreachable")
	}
}

func (com Com) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	switch com.Tag {
	case AndComTag:
		i := input.(map[string]interface{})
		a := i["a"].(map[string]interface{})
		b := i["b"].(map[string]interface{})
		aTrue := a["true"].(<-chan Unit)
		aFalse := a["false"].(<-chan Unit)
		bTrue := b["true"].(<-chan Unit)
		bFalse := b["false"].(<-chan Unit)

		o := output.(map[string]interface{})
		oTrue := o["true"].(chan<- Unit)
		oFalse := o["false"].(chan<- Unit)

		runAnd(aTrue, aFalse, bTrue, bFalse, oTrue, oFalse, quit)
	case MergeComTag:
		inputStruct := input.(map[string]interface{})
		runMerge(com.Extra.(DType), inputStruct["a"], inputStruct["b"], output, quit)
	case SplitComTag:
		outputStruct := output.(map[string]interface{})
		runSplit(com.Extra.(DType), input, outputStruct["a"], outputStruct["b"], quit)
	case ChooseComTag:
		i := input.(map[string]interface{})
		iA := i["a"].(<-chan Unit)
		iB := i["b"].(<-chan Unit)
		ready := i["ready"].(<-chan Unit)

		o := output.(map[string]interface{})
		oA := o["a"].(chan<- Unit)
		oB := o["b"].(chan<- Unit)

		runChoose(iA, iB, ready, oA, oB, quit)
	case ConstComTag:
		extra := com.Extra.(ConstComExtra)
		runConst(extra.Type, extra.Val, input.(<-chan Unit), output, quit)
	case IComTag:
		runI(com.Extra.(DType), input, output, quit)
	case SelectComTag:
		extra := com.Extra.(SelectExtra)
		runI(extra.Fields[extra.FieldName], input.(map[string]interface{})[extra.FieldName], output, quit)
	case UnselectComTag:
		extra := com.Extra.(UnselectExtra)
		runI(extra.FieldType, input, output.(map[string]interface{})[extra.FieldName], quit)
	case CompositeComTag:
		runComposite(com.Extra.(CompositeComExtra), input, output, quit)
	}
}

func runAnd(aTrue, aFalse, bTrue, bFalse <-chan Unit, oTrue, oFalse chan<- Unit, quit <-chan struct{}) {
	var a, b bool
	for {
		select {
		case <-aTrue:
			a = true
		case <-aFalse:
			a = false
		case <-quit:
			return
		}
		select {
		case <-bTrue:
			b = true
		case <-bFalse:
			b = false
		case <-quit:
			return
		}
		if a && b {
			oTrue <- Unit{}
		} else {
			oFalse <- Unit{}
		}
	}
}

func runMerge(theType DType, inputA, inputB interface{}, output interface{}, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		a, b := inputA.(<-chan Unit), inputB.(<-chan Unit)
		o := output.(chan<- Unit)
		for {
			select {
			case <-a:
				o <- Unit{}
			case <-b:
				o <- Unit{}
			case <-quit:
				return
			}
		}
	case DTypeTypeTag:
		a, b := inputA.(<-chan DType), inputB.(<-chan DType)
		o := output.(chan<- DType)
		for {
			select {
			case v := <-a:
				o <- v
			case v := <-b:
				o <- v
			case <-quit:
				return
			}
		}
	case ComTypeTag:
		a, b := inputA.(<-chan Com), inputB.(<-chan Com)
		o := output.(chan<- Com)
		for {
			select {
			case v := <-a:
				o <- v
			case v := <-b:
				o <- v
			case <-quit:
				return
			}
		}
	case StructTypeTag:
		a, b := inputA.(map[string]interface{}), inputB.(map[string]interface{})
		o := output.(map[string]interface{})
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			go runMerge(fieldType, a[fieldName], b[fieldName], o[fieldName], quit)
		}
	}
}

func merge(theType DType) Com {
	return Com{MergeComTag, theType}
}

func runSplit(theType DType, input interface{}, outputA, outputB interface{}, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		i := input.(<-chan Unit)
		a, b := outputA.(chan<- Unit), outputB.(chan<- Unit)
		for {
			select {
			case <-i:
				a <- Unit{}
				b <- Unit{}
			case <-quit:
				return
			}
		}
	case DTypeTypeTag:
		i := input.(<-chan DType)
		a, b := outputA.(chan<- DType), outputB.(chan<- DType)
		for {
			select {
			case v := <-i:
				a <- v
				b <- v
			case <-quit:
				return
			}
		}
	case ComTypeTag:
		i := input.(<-chan Com)
		a, b := outputA.(chan<- Com), outputB.(chan<- Com)
		for {
			select {
			case v := <-i:
				a <- v
				b <- v
			case <-quit:
				return
			}
		}
	case StructTypeTag:
		i := input.(map[string]interface{})
		a, b := outputA.(map[string]interface{}), outputB.(map[string]interface{})
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			go runSplit(fieldType, i[fieldName], a[fieldName], b[fieldName], quit)
		}
	}
}

func split(theType DType) Com {
	return Com{SplitComTag, theType}
}

func runChoose(iA, iB, ready <-chan Unit, oA, oB chan<- Unit, quit <-chan struct{}) {
	for {
		select {
		case <-ready:
		case <-quit:
			return
		}
		select {
		case <-iA:
			oA <- Unit{}
		case <-iB:
			oB <- Unit{}
		case <-quit:
			return
		}
	}
}

type inputChanSource struct {
	Units []<-chan Unit
	Types []<-chan DType
	Coms  []<-chan Com
}

type outputChanSource struct {
	Units []chan<- Unit
	Types []chan<- DType
	Coms  []chan<- Com
}

func makeInputChanSource(n CompositeComChanSourceN) (ret inputChanSource) {
	ret.Units = make([]<-chan Unit, n.Units)
	ret.Types = make([]<-chan DType, n.Types)
	ret.Coms = make([]<-chan Com, n.Coms)
	return
}

func makeOutputChanSource(n CompositeComChanSourceN) (ret outputChanSource) {
	ret.Units = make([]chan<- Unit, n.Units)
	ret.Types = make([]chan<- DType, n.Types)
	ret.Coms = make([]chan<- Com, n.Coms)
	return
}

func putInputChans(dType DType, chanMap interface{}, input interface{}, chans inputChanSource) {
	switch dType.Tag {
	case UnitTypeTag:
		chans.Units[chanMap.(int)] = input.(<-chan Unit)
	case DTypeTypeTag:
		chans.Types[chanMap.(int)] = input.(<-chan DType)
	case ComTypeTag:
		chans.Coms[chanMap.(int)] = input.(<-chan Com)
	case StructTypeTag:
		for fieldName, fieldType := range dType.Extra.(map[string]DType) {
			putInputChans(fieldType, chanMap.(map[string]interface{})[fieldName], input.(map[string]interface{})[fieldName], chans)
		}
	}
}

func putOutputChans(dType DType, chanMap interface{}, input interface{}, chans outputChanSource) {
	switch dType.Tag {
	case UnitTypeTag:
		chans.Units[chanMap.(int)] = input.(chan<- Unit)
	case DTypeTypeTag:
		chans.Types[chanMap.(int)] = input.(chan<- DType)
	case ComTypeTag:
		chans.Coms[chanMap.(int)] = input.(chan<- Com)
	case StructTypeTag:
		for fieldName, fieldType := range dType.Extra.(map[string]DType) {
			putOutputChans(fieldType, chanMap.(map[string]interface{})[fieldName], input.(map[string]interface{})[fieldName], chans)
		}
	}
}

func getInput(dType DType, chanMap interface{}, chans inputChanSource) interface{} {
	switch dType.Tag {
	case UnitTypeTag:
		return chans.Units[chanMap.(int)]
	case DTypeTypeTag:
		return chans.Types[chanMap.(int)]
	case ComTypeTag:
		return chans.Coms[chanMap.(int)]
	case StructTypeTag:
		input := make(map[string]interface{})
		for fieldName, fieldType := range dType.Extra.(map[string]DType) {
			input[fieldName] = getInput(fieldType, chanMap.(map[string]interface{})[fieldName], chans)
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
	case DTypeTypeTag:
		return chans.Types[chanMap.(int)]
	case ComTypeTag:
		return chans.Coms[chanMap.(int)]
	case StructTypeTag:
		output := make(map[string]interface{})
		for fieldName, fieldType := range dType.Extra.(map[string]DType) {
			output[fieldName] = getOutput(fieldType, chanMap.(map[string]interface{})[fieldName], chans)
		}
		return output
	default:
		panic("Unreachable")
	}
}

type constDTypeEntry struct {
	Chan chan<- DType
	Val  DType
}

type constComEntry struct {
	Chan chan<- Com
	Val  Com
}

func putConstEntries(units *[]chan<- Unit, types *[]constDTypeEntry, coms *[]constComEntry, theType DType, val interface{}, output interface{}) {
	switch theType.Tag {
	case UnitTypeTag:
		constVal := val.(ConstVal)
		if constVal.P {
			*units = append(*units, output.(chan<- Unit))
		}
	case DTypeTypeTag:
		constVal := val.(ConstVal)
		if constVal.P {
			*types = append(*types, constDTypeEntry{output.(chan<- DType), constVal.Val.(DType)})
		}
	case ComTypeTag:
		constVal := val.(ConstVal)
		if constVal.P {
			*coms = append(*coms, constComEntry{output.(chan<- Com), constVal.Val.(Com)})
		}
	case StructTypeTag:
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			putConstEntries(units, types, coms, fieldType, val.(map[string]interface{})[fieldName], output.(map[string]interface{})[fieldName])
		}
	}
}

func runConst(theType DType, val interface{}, input <-chan Unit, output interface{}, quit <-chan struct{}) {
	var units []chan<- Unit
	var types []constDTypeEntry
	var coms []constComEntry

	putConstEntries(&units, &types, &coms, theType, val, output)

	for {
		select {
		case <-input:
			for _, entry := range units {
				entry <- Unit{}
			}
			for _, entry := range types {
				entry.Chan <- entry.Val
			}
			for _, entry := range coms {
				entry.Chan <- entry.Val
			}
		case <-quit:
			return
		}
	}
}

func makeConst(theType DType, val interface{}) Com {
	return Com{ConstComTag, ConstComExtra{theType, val}}
}

func runI(theType DType, input interface{}, output interface{}, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		i := input.(<-chan Unit)
		o := input.(chan<- Unit)
		select {
		case <-i:
			o <- Unit{}
		case <-quit:
			return
		}
	case DTypeTypeTag:
		i := input.(<-chan DType)
		o := input.(chan<- DType)
		select {
		case v := <-i:
			o <- v
		case <-quit:
			return
		}
	case ComTypeTag:
		i := input.(<-chan Com)
		o := input.(chan<- Com)
		select {
		case v := <-i:
			o <- v
		case <-quit:
			return
		}
	case StructTypeTag:
		i := input.(map[string]interface{})
		o := output.(map[string]interface{})
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			go runI(fieldType, i[fieldName], o[fieldName], quit)
		}
	}
}

func makeI(theType DType) Com {
	return Com{IComTag, theType}
}

func makeSelect(fieldName string, fields map[string]DType) Com {
	return Com{SelectComTag, SelectExtra{fields, fieldName}}
}

func makeUnselect(fieldName string, fieldType DType) Com {
	return Com{UnselectComTag, UnselectExtra{fieldName, fieldType}}
}

func runComposite(extra CompositeComExtra, input interface{}, output interface{}, quit <-chan struct{}) {
	inChans := makeInputChanSource(extra.InputChanN)
	outChans := makeOutputChanSource(extra.OutputChanN)

	putInputChans(extra.InputType, extra.InputMap, input, inChans)
	putOutputChans(extra.OutputType, extra.OutputMap, output, outChans)

	for i := 0; i < extra.InnerChanN.Units; i++ {
		theChan := make(chan Unit, 1)
		inChans.Units[i] = (<-chan Unit)(theChan)
		outChans.Units[i] = chan<- Unit(theChan)
	}
	for i := 0; i < extra.InnerChanN.Types; i++ {
		theChan := make(chan DType, 1)
		inChans.Types[i] = (<-chan DType)(theChan)
		outChans.Types[i] = chan<- DType(theChan)
	}
	for i := 0; i < extra.InnerChanN.Coms; i++ {
		theChan := make(chan Com, 1)
		inChans.Coms[i] = (<-chan Com)(theChan)
		outChans.Coms[i] = chan<- Com(theChan)
	}

	for _, comEntry := range extra.ComEntries {
		input := getInput(comEntry.InputType(), comEntry.InputMap, inChans)
		output := getOutput(comEntry.OutputType(), comEntry.OutputMap, outChans)
		go comEntry.Run(input, output, quit)
	}
}

func makeMaps(map0, map1 *interface{}, chanN *CompositeComChanSourceN, theType DType) {
	switch theType.Tag {
	case UnitTypeTag:
		*map0 = chanN.Units
		*map1 = chanN.Units
		chanN.Units++
	case DTypeTypeTag:
		*map0 = chanN.Types
		*map1 = chanN.Types
		chanN.Types++
	case ComTypeTag:
		*map0 = chanN.Coms
		*map1 = chanN.Coms
		chanN.Coms++
	case StructTypeTag:
		map0Val := make(map[string]interface{})
		*map0 = map0Val

		map1Val := make(map[string]interface{})
		*map1 = map1Val

		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			var fieldMap0 interface{}
			var fieldMap1 interface{}

			makeMaps(&fieldMap0, &fieldMap1, chanN, fieldType)

			map0Val[fieldName] = fieldMap0
			map1Val[fieldName] = fieldMap1
		}
	}
}

// len(coms) > 0
func pipe(coms []Com) Com {
	var extra CompositeComExtra

	extra.InputType = coms[0].InputType()
	extra.OutputType = coms[len(coms)-1].OutputType()

	extra.ComEntries = make([]CompositeComEntry, len(coms))
	for i, com := range coms {
		extra.ComEntries[i].Com = com
	}

	for i := 0; i < len(extra.ComEntries)-1; i++ {
		//TODO: Check types
		makeMaps(&extra.ComEntries[i].OutputMap, &extra.ComEntries[i+1].InputMap, &extra.InputChanN, extra.ComEntries[i].OutputType())
	}

	extra.OutputChanN = extra.InputChanN
	extra.InnerChanN = extra.InputChanN

	makeMaps(&extra.InputMap, &extra.ComEntries[0].InputMap, &extra.InputChanN, extra.InputType)
	makeMaps(&extra.OutputMap, &extra.ComEntries[len(coms)-1].OutputMap, &extra.OutputChanN, extra.OutputType)

	return Com{CompositeComTag, extra}
}

func main() {
	BoolTypeFields["true"] = UnitType
	BoolTypeFields["false"] = UnitType

	andComInputTypeFields["a"] = BoolType
	andComInputTypeFields["b"] = BoolType

	chooseComInputTypeFields["a"] = UnitType
	chooseComInputTypeFields["b"] = UnitType
	chooseComInputTypeFields["ready"] = UnitType

	chooseComOutputTypeFields["a"] = UnitType
	chooseComOutputTypeFields["b"] = UnitType

	com := pipe([]Com{split(BoolType), AndCom})

	inputI, inputO := makeIOChans(com.InputType())
	outputI, outputO := makeIOChans(com.OutputType())

	quit := make(chan struct{})
	defer close(quit)

	go com.Run(inputI, outputO, quit)

	inputO.(map[string]interface{})["true"].(chan<- Unit) <- Unit{}
	select {
	case <-outputI.(map[string]interface{})["true"].(<-chan Unit):
		fmt.Println(true)
	case <-outputI.(map[string]interface{})["false"].(<-chan Unit):
		fmt.Println(false)
	}

	inputO.(map[string]interface{})["false"].(chan<- Unit) <- Unit{}
	select {
	case <-outputI.(map[string]interface{})["true"].(<-chan Unit):
		fmt.Println(true)
	case <-outputI.(map[string]interface{})["false"].(<-chan Unit):
		fmt.Println(false)
	}
}
