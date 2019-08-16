package main

import "fmt"

type DTypeTag int

const (
	UnitTypeTag = DTypeTag(iota)
	SyntaxTypeTag
	GenComTypeTag
	StructTypeTag
)

type Unit struct{}

type SyntaxTag int

const (
	StringSyntaxTag = SyntaxTag(iota)
	LolSyntaxTag
	MCallSyntaxTag
	QuotedComSyntaxTag
)

type String string

type Lol [][]Syntax

type MCall struct {
	Macro Syntax /* String or MCall */
	Arg   Syntax /* String or Lol */
}

type QuotedCom GenCom

type Syntax struct {
	Tag   SyntaxTag
	Extra interface{}
}

type GenCom func(inputType DType) Com

type Struct map[string]interface{}

type DType struct {
	Tag   DTypeTag
	Extra interface{}
}

var UnitType = DType{UnitTypeTag, nil}

var SyntaxType = DType{SyntaxTypeTag, nil}

var GenComType = DType{GenComTypeTag, nil}

func makeStructType(fields map[string]DType) DType {
	return DType{StructTypeTag, fields}
}

var BoolTypeFields map[string]DType = make(map[string]DType, 2)
var BoolType DType = makeStructType(BoolTypeFields)

func writeBool(output interface{}, val bool) {
	if val {
		output.(Struct)["true"].(chan<- Unit) <- Unit{}
	} else {
		output.(Struct)["false"].(chan<- Unit) <- Unit{}
	}
}

func readBool(input interface{}) bool {
	select {
	case <-input.(Struct)["true"].(<-chan Unit):
		return true
	case <-input.(Struct)["false"].(<-chan Unit):
		return false
	}
}

func makeIOChans(theType DType) (input, output interface{}) {
	switch theType.Tag {
	case UnitTypeTag:
		theChan := make(chan Unit, 1)
		input = (<-chan Unit)(theChan)
		output = chan<- Unit(theChan)
	case SyntaxTypeTag:
		theChan := make(chan Syntax, 1)
		input = (<-chan Syntax)(theChan)
		output = chan<- Syntax(theChan)
	case GenComTypeTag:
		theChan := make(chan GenCom, 1)
		input = (<-chan GenCom)(theChan)
		output = chan<- GenCom(theChan)
	case StructTypeTag:
		inputMap := make(Struct)
		outputMap := make(Struct)
		input = inputMap
		output = outputMap
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			inputMap[fieldName], outputMap[fieldName] = makeIOChans(fieldType)
		}
	}
	return
}

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
	i := input.(Struct)
	a := i["a"].(Struct)
	b := i["b"].(Struct)
	aTrue := a["true"].(<-chan Unit)
	aFalse := a["false"].(<-chan Unit)
	bTrue := b["true"].(<-chan Unit)
	bFalse := b["false"].(<-chan Unit)

	o := output.(Struct)
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
	case SyntaxTypeTag:
		a, b := inputA.(<-chan Syntax), inputB.(<-chan Syntax)
		o := output.(chan<- Syntax)
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
		a, b := inputA.(Struct), inputB.(Struct)
		o := output.(Struct)
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			go runMerge(fieldType, a[fieldName], b[fieldName], o[fieldName], quit)
		}
	}
}

func (com MergeCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	inputStruct := input.(Struct)
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
	case SyntaxTypeTag:
		i := input.(<-chan Syntax)
		a, b := outputA.(chan<- Syntax), outputB.(chan<- Syntax)
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
		i := input.(Struct)
		a, b := outputA.(Struct), outputB.(Struct)
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			go runSplit(fieldType, i[fieldName], a[fieldName], b[fieldName], quit)
		}
	}
}

func (com SplitCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	outputStruct := output.(Struct)
	runSplit(DType(com), input, outputStruct["a"], outputStruct["b"], quit)
}

func (com ChooseCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	i := input.(Struct)
	iA := i["a"].(<-chan Unit)
	iB := i["b"].(<-chan Unit)
	ready := i["ready"].(<-chan Unit)

	o := output.(Struct)
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

type constSyntaxEntry struct {
	Chan chan<- Syntax
	Val  Syntax
}

type constGenComEntry struct {
	Chan chan<- GenCom
	Val  GenCom
}

func putConstEntries(units *[]chan<- Unit, syntaxen *[]constSyntaxEntry, genComs *[]constGenComEntry, theType DType, val interface{}, output interface{}) {
	switch theType.Tag {
	case UnitTypeTag:
		constVal := val.(ConstVal)
		if constVal.P {
			*units = append(*units, output.(chan<- Unit))
		}
	case SyntaxTypeTag:
		constVal := val.(ConstVal)
		if constVal.P {
			*syntaxen = append(*syntaxen, constSyntaxEntry{output.(chan<- Syntax), constVal.Val.(Syntax)})
		}
	case GenComTypeTag:
		constVal := val.(ConstVal)
		if constVal.P {
			*genComs = append(*genComs, constGenComEntry{output.(chan<- GenCom), constVal.Val.(GenCom)})
		}
	case StructTypeTag:
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			putConstEntries(units, syntaxen, genComs, fieldType, val.(Struct)[fieldName], output.(Struct)[fieldName])
		}
	}
}

func (com ConstCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	i := input.(<-chan Unit)

	var units []chan<- Unit
	var syntaxen []constSyntaxEntry
	var genComs []constGenComEntry

	putConstEntries(&units, &syntaxen, &genComs, com.Type, com.Val, output)

	for {
		select {
		case <-i:
			for _, entry := range units {
				entry <- Unit{}
			}
			for _, entry := range syntaxen {
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
	case SyntaxTypeTag:
		i := input.(<-chan Syntax)
		o := output.(chan<- Syntax)
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
		i := input.(Struct)
		o := output.(Struct)
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
	case SyntaxTypeTag:
		i := input.(<-chan Syntax)
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
		i := input.(Struct)
		for fieldName, fieldType := range theType.Extra.(map[string]DType) {
			go runSink(fieldType, i[fieldName], quit)
		}
	}
}

func (com SinkCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	runSink(DType(com), input, quit)
}

func (com SelectCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	i := input.(Struct)
	for fieldName, fieldType := range com.Fields {
		if fieldName == com.FieldName {
			go runI(fieldType, i[fieldName], output, quit)
		} else {
			go runSink(fieldType, i[fieldName], quit)
		}
	}
}
func (com DeselectCom) Run(input interface{}, output interface{}, quit <-chan struct{}) {
	runI(com.FieldType, input, output.(Struct)[com.FieldName], quit)
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

func makeMaps(map0, map1 *interface{}, chanN *CompositeComChanSourceN, theType DType) {
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
		map0Val := make(Struct)
		*map0 = map0Val

		map1Val := make(Struct)
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

func genAnd(inputType DType) Com { /* TODO: Check inputType */ return AndCom{} }

func genMerge(inputType DType) Com { return MergeCom(inputType) }

func genSplit(inputType DType) Com { return SplitCom(inputType) }

func genChoose(inputType DType) Com { /* TODO: Check inputType */ return ChooseCom{} }

func genConst(outputType DType, val interface{}) GenCom {
	return func(inputType DType) Com { /* TODO: Check inputType */ return ConstCom{outputType, val} }
}

func genI(inputType DType) Com { return ICom(inputType) }

func genSink(inputType DType) Com { return SinkCom(inputType) }

func genSelect(fieldName string) GenCom {
	return func(inputType DType) Com {
		if inputType.Tag != StructTypeTag {
			panic("Type error")
		}
		return SelectCom{inputType.Extra.(map[string]DType), fieldName}
	}
}

func genDeselect(fieldName string) GenCom {
	return func(inputType DType) Com { return DeselectCom{fieldName, inputType} }
}

func genPipe(genComs []GenCom) GenCom {
	return func(inputType DType) Com {
		coms := make([]Com, len(genComs))
		for i, genCom := range genComs {
			com := genCom(inputType)
			coms[i] = com
			inputType = com.OutputType()
		}
		return pipe(coms)
	}
}

func init() {
	BoolTypeFields["true"] = UnitType
	BoolTypeFields["false"] = UnitType

	andComInputTypeFields["a"] = BoolType
	andComInputTypeFields["b"] = BoolType

	chooseComInputTypeFields["a"] = UnitType
	chooseComInputTypeFields["b"] = UnitType
	chooseComInputTypeFields["ready"] = UnitType

	chooseComOutputTypeFields["a"] = UnitType
	chooseComOutputTypeFields["b"] = UnitType
}

func main() {
	com := genPipe([]GenCom{genI, genSplit, genAnd})(BoolType)

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
