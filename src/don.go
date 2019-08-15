package main

import "fmt"

type DTypeTag int

const (
	UnitTypeTag = DTypeTag(iota)
	StringTypeTag
	LoLTypeTag
	GenComTypeTag
	StructTypeTag
)

type Unit struct{}

type String string

type LoL [][]interface{}

type DType struct {
	Tag   DTypeTag
	Extra interface{}
}

var UnitType = DType{UnitTypeTag, nil}

var StringType = DType{StringTypeTag, nil}

var LoLType = DType{LoLTypeTag, nil}

func makeStructType(fields map[string]DType) DType {
	return DType{StructTypeTag, fields}
}

func makeGenComType(inputType, outputType DType) DType {
	return DType{GenComTypeTag, [2]DType{inputType, outputType}}
}

var BoolTypeFields map[string]DType = make(map[string]DType, 2)
var BoolType DType = makeStructType(BoolTypeFields)

func writeBool(output interface{}, val bool) {
	if val {
		output.(map[string]interface{})["true"].(chan<- Unit) <- Unit{}
	} else {
		output.(map[string]interface{})["false"].(chan<- Unit) <- Unit{}
	}
}

func readBool(input interface{}) bool {
	select {
	case <-input.(map[string]interface{})["true"].(<-chan Unit):
		return true
	case <-input.(map[string]interface{})["false"].(<-chan Unit):
		return false
	}
}

func makeIOChans(theType DType) (input, output interface{}) {
	switch theType.Tag {
	case UnitTypeTag:
		theChan := make(chan Unit, 1)
		input = (<-chan Unit)(theChan)
		output = chan<- Unit(theChan)
	case StringTypeTag:
		theChan := make(chan String, 1)
		input = (<-chan String)(theChan)
		output = chan<- String(theChan)
	case LoLTypeTag:
		theChan := make(chan LoL, 1)
		input = (<-chan LoL)(theChan)
		output = chan<- LoL(theChan)
	case GenComTypeTag:
		theChan := make(chan GenCom, 1)
		input = (<-chan GenCom)(theChan)
		output = chan<- GenCom(theChan)
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

type GenCom func(inputType DType) Com

type Com interface {
	Run(input interface{}, output interface{}, quit <-chan struct{})
	InputType() DType
	OutputType() DType
}

type AndCom struct{}
type MergeCom DType
type SplitCom DType
type ChooseCom struct{}

type ConstVal struct {
	P   bool
	Val interface{}
}

type ConstCom struct {
	Type DType
	Val  interface{}
}

type ICom DType
type SinkCom DType

type SelectCom struct {
	Fields    map[string]DType
	FieldName string
}

type DeselectCom struct {
	FieldName string
	FieldType DType
}

type CompositeComChanSourceN struct {
	Units, Strings, LoLs, GenComs int
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

var andComInputTypeFields map[string]DType = make(map[string]DType, 2)
var andComInputType DType = makeStructType(andComInputTypeFields)

var chooseComInputTypeFields map[string]DType = make(map[string]DType, 3)
var chooseComInputType DType = makeStructType(chooseComInputTypeFields)

var chooseComOutputTypeFields map[string]DType = make(map[string]DType, 2)
var chooseComOutputType DType = makeStructType(chooseComOutputTypeFields)

func (com AndCom) InputType() DType {
	return andComInputType
}
func (com MergeCom) InputType() DType {
	theType := DType(com)
	fields := make(map[string]DType, 2)
	fields["a"] = theType
	fields["b"] = theType
	return DType{StructTypeTag, fields}
}
func (com SplitCom) InputType() DType {
	return DType(com)
}
func (com ChooseCom) InputType() DType {
	return chooseComInputType
}
func (com ConstCom) InputType() DType {
	return UnitType
}
func (com ICom) InputType() DType {
	return DType(com)
}
func (com SinkCom) InputType() DType {
	return DType(com)
}
func (com SelectCom) InputType() DType {
	return makeStructType(com.Fields)
}
func (com DeselectCom) InputType() DType {
	return com.FieldType
}
func (com CompositeCom) InputType() DType {
	return com.TheInputType
}

func (com AndCom) OutputType() DType {
	return BoolType
}
func (com MergeCom) OutputType() DType {
	return DType(com)
}
func (com SplitCom) OutputType() DType {
	theType := DType(com)
	fields := make(map[string]DType, 2)
	fields["a"] = theType
	fields["b"] = theType
	return DType{StructTypeTag, fields}
}
func (com ChooseCom) OutputType() DType {
	return chooseComOutputType
}
func (com ConstCom) OutputType() DType {
	return com.Type
}
func (com ICom) OutputType() DType {
	return DType(com)
}
func (com SinkCom) OutputType() DType {
	return makeStructType(make(map[string]DType, 0))
}
func (com SelectCom) OutputType() DType {
	return com.Fields[com.FieldName]
}
func (com DeselectCom) OutputType() DType {
	fields := make(map[string]DType, 1)
	fields[com.FieldName] = com.FieldType
	return makeStructType(fields)
}
func (com CompositeCom) OutputType() DType {
	return com.TheOutputType
}

// Each chan (except quit) corresponds to exactly one input
func (com AndCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
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

	var aVal, bVal bool
	for {
		select {
		case <-aTrue:
			aVal = true
		case <-aFalse:
			aVal = false
		case <-quit:
			return
		}
		select {
		case <-bTrue:
			bVal = true
		case <-bFalse:
			bVal = false
		case <-quit:
			return
		}
		if aVal && bVal {
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
	case StringTypeTag:
		a, b := inputA.(<-chan String), inputB.(<-chan String)
		o := output.(chan<- String)
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
	case LoLTypeTag:
		a, b := inputA.(<-chan LoL), inputB.(<-chan LoL)
		o := output.(chan<- LoL)
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
	case GenComTypeTag:
		a, b := inputA.(<-chan GenCom), inputB.(<-chan GenCom)
		o := output.(chan<- GenCom)
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

func (com MergeCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	inputStruct := input.(map[string]interface{})
	runMerge(DType(com), inputStruct["a"], inputStruct["b"], output, quit)
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
	case StringTypeTag:
		i := input.(<-chan String)
		a, b := outputA.(chan<- String), outputB.(chan<- String)
		for {
			select {
			case v := <-i:
				a <- v
				b <- v
			case <-quit:
				return
			}
		}
	case LoLTypeTag:
		i := input.(<-chan LoL)
		a, b := outputA.(chan<- LoL), outputB.(chan<- LoL)
		for {
			select {
			case v := <-i:
				a <- v
				b <- v
			case <-quit:
				return
			}
		}
	case GenComTypeTag:
		i := input.(<-chan GenCom)
		a, b := outputA.(chan<- GenCom), outputB.(chan<- GenCom)
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

func (com SplitCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	outputStruct := output.(map[string]interface{})
	runSplit(DType(com), input, outputStruct["a"], outputStruct["b"], quit)
}

func (com ChooseCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	i := input.(map[string]interface{})
	iA := i["a"].(<-chan Unit)
	iB := i["b"].(<-chan Unit)
	ready := i["ready"].(<-chan Unit)

	o := output.(map[string]interface{})
	oA := o["a"].(chan<- Unit)
	oB := o["b"].(chan<- Unit)

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

type constStringEntry struct {
	Chan chan<- String
	Val  String
}

type constLoLEntry struct {
	Chan chan<- LoL
	Val  LoL
}

type constGenComEntry struct {
	Chan chan<- GenCom
	Val  GenCom
}

func putConstEntries(units *[]chan<- Unit, strings *[]constStringEntry, loLs *[]constLoLEntry, genComs *[]constGenComEntry, theType DType, val interface{}, output interface{}) {
	switch theType.Tag {
	case UnitTypeTag:
		constVal := val.(ConstVal)
		if constVal.P {
			*units = append(*units, output.(chan<- Unit))
		}
	case StringTypeTag:
		constVal := val.(ConstVal)
		if constVal.P {
			*strings = append(*strings, constStringEntry{output.(chan<- String), constVal.Val.(String)})
		}
	case LoLTypeTag:
		constVal := val.(ConstVal)
		if constVal.P {
			*loLs = append(*loLs, constLoLEntry{output.(chan<- LoL), constVal.Val.(LoL)})
		}
	case GenComTypeTag:
		constVal := val.(ConstVal)
		if constVal.P {
			*genComs = append(*genComs, constGenComEntry{output.(chan<- GenCom), constVal.Val.(GenCom)})
		}
	case StructTypeTag:
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			putConstEntries(units, strings, loLs, genComs, fieldType, val.(map[string]interface{})[fieldName], output.(map[string]interface{})[fieldName])
		}
	}
}

func (com ConstCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	i := input.(<-chan Unit)

	var units []chan<- Unit
	var strings []constStringEntry
	var loLs []constLoLEntry
	var genComs []constGenComEntry

	putConstEntries(&units, &strings, &loLs, &genComs, com.Type, com.Val, output)

	for {
		select {
		case <-i:
			for _, entry := range units {
				entry <- Unit{}
			}
			for _, entry := range strings {
				entry.Chan <- entry.Val
			}
			for _, entry := range loLs {
				entry.Chan <- entry.Val
			}
			for _, entry := range genComs {
				entry.Chan <- entry.Val
			}
		case <-quit:
			return
		}
	}
}

func runI(theType DType, input interface{}, output interface{}, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		i := input.(<-chan Unit)
		o := output.(chan<- Unit)
		for {
			select {
			case <-i:
				o <- Unit{}
			case <-quit:
				return
			}
		}
	case StringTypeTag:
		i := input.(<-chan String)
		o := output.(chan<- String)
		for {
			select {
			case v := <-i:
				o <- v
			case <-quit:
				return
			}
		}
	case LoLTypeTag:
		i := input.(<-chan LoL)
		o := output.(chan<- LoL)
		for {
			select {
			case v := <-i:
				o <- v
			case <-quit:
				return
			}
		}
	case GenComTypeTag:
		i := input.(<-chan GenCom)
		o := output.(chan<- GenCom)
		for {
			select {
			case v := <-i:
				o <- v
			case <-quit:
				return
			}
		}
	case StructTypeTag:
		i := input.(map[string]interface{})
		o := output.(map[string]interface{})
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			go runI(fieldType, i[fieldName], o[fieldName], quit)
		}
	}
}

func (com ICom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	runI(DType(com), input, output, quit)
}

func runSink(theType DType, input interface{}, quit <-chan struct{}) {
	switch theType.Tag {
	case UnitTypeTag:
		i := input.(<-chan Unit)
		for {
			select {
			case <-i:
			case <-quit:
				return
			}
		}
	case StringTypeTag:
		i := input.(<-chan String)
		for {
			select {
			case <-i:
			case <-quit:
				return
			}
		}
	case LoLTypeTag:
		i := input.(<-chan LoL)
		for {
			select {
			case <-i:
			case <-quit:
				return
			}
		}
	case GenComTypeTag:
		i := input.(<-chan GenCom)
		for {
			select {
			case <-i:
			case <-quit:
				return
			}
		}
	case StructTypeTag:
		i := input.(map[string]interface{})
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			go runSink(fieldType, i[fieldName], quit)
		}
	}
}

func (com SinkCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	runSink(DType(com), input, quit)
}

func (com SelectCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	i := input.(map[string]interface{})
	for fieldName, fieldType := range com.Fields {
		if fieldName == com.FieldName {
			go runI(fieldType, i[fieldName], output, quit)
		} else {
			go runSink(fieldType, i[fieldName], quit)
		}
	}
}
func (com DeselectCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	runI(com.FieldType, input, output.(map[string]interface{})[com.FieldName], quit)
}

type inputChanSource struct {
	Units   []<-chan Unit
	Strings []<-chan String
	LoLs    []<-chan LoL
	GenComs []<-chan GenCom
}

type outputChanSource struct {
	Units   []chan<- Unit
	Strings []chan<- String
	LoLs    []chan<- LoL
	GenComs []chan<- GenCom
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
	case StringTypeTag:
		chans.Strings[chanMap.(int)] = input.(<-chan String)
	case LoLTypeTag:
		chans.LoLs[chanMap.(int)] = input.(<-chan LoL)
	case GenComTypeTag:
		chans.GenComs[chanMap.(int)] = input.(<-chan GenCom)
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
	case StringTypeTag:
		chans.Strings[chanMap.(int)] = input.(chan<- String)
	case LoLTypeTag:
		chans.LoLs[chanMap.(int)] = input.(chan<- LoL)
	case GenComTypeTag:
		chans.GenComs[chanMap.(int)] = input.(chan<- GenCom)
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
	case StringTypeTag:
		return chans.Strings[chanMap.(int)]
	case LoLTypeTag:
		return chans.LoLs[chanMap.(int)]
	case GenComTypeTag:
		return chans.GenComs[chanMap.(int)]
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
	case StringTypeTag:
		return chans.Strings[chanMap.(int)]
	case LoLTypeTag:
		return chans.LoLs[chanMap.(int)]
	case GenComTypeTag:
		return chans.GenComs[chanMap.(int)]
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

func makeMaps(map0, map1 *interface{}, chanN *CompositeComChanSourceN, theType DType) {
	switch theType.Tag {
	case UnitTypeTag:
		*map0 = chanN.Units
		*map1 = chanN.Units
		chanN.Units++
	case StringTypeTag:
		*map0 = chanN.Strings
		*map1 = chanN.Strings
		chanN.Strings++
	case LoLTypeTag:
		*map0 = chanN.LoLs
		*map1 = chanN.LoLs
		chanN.LoLs++
	case GenComTypeTag:
		*map0 = chanN.GenComs
		*map1 = chanN.GenComs
		chanN.GenComs++
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
func pipe(coms []Com) (ret CompositeCom) {
	ret.TheInputType = coms[0].InputType()
	ret.TheOutputType = coms[len(coms)-1].OutputType()

	ret.ComEntries = make([]CompositeComEntry, len(coms))
	for i, com := range coms {
		ret.ComEntries[i].Com = com
	}

	for i := 0; i < len(ret.ComEntries)-1; i++ {
		//TODO: Check types
		makeMaps(&ret.ComEntries[i].OutputMap, &ret.ComEntries[i+1].InputMap, &ret.InputChanN, ret.ComEntries[i].OutputType())
	}

	ret.OutputChanN = ret.InputChanN
	ret.InnerChanN = ret.InputChanN

	makeMaps(&ret.InputMap, &ret.ComEntries[0].InputMap, &ret.InputChanN, ret.TheInputType)
	makeMaps(&ret.OutputMap, &ret.ComEntries[len(coms)-1].OutputMap, &ret.OutputChanN, ret.TheOutputType)

	return
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

	com := pipe([]Com{ICom(BoolType), SplitCom(BoolType), AndCom{}})

	inputI, inputO := makeIOChans(com.InputType())
	outputI, outputO := makeIOChans(com.OutputType())

	quit := make(chan struct{})
	defer close(quit)

	go com.Run(inputI, outputO, quit)

	writeBool(inputO, true)
	fmt.Println(readBool(outputI))

	writeBool(inputO, true)
	fmt.Println(readBool(outputI))

	writeBool(inputO, false)
	fmt.Println(readBool(outputI))
}
