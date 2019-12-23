package coms

import . "don/core"

type MergeCom struct{}

func (MergeCom) OutputType(inputType DType) DType {
	if inputType.Lvl != NormalTypeLvl {
		return inputType
	}
	if inputType.Tag != StructTypeTag {
		return ImpossibleType
	}

	ret := UnknownType
	for _, subType := range inputType.Fields {
		ret = MergeTypes(ret, subType)
	}
	return ret
}

func (MergeCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	com := CompositeCom{
		Coms: make([]CompositeComEntry, len(inputType.Fields)),
		InputMap: SignalReaderIdTree{
			ParentP:  true,
			Children: make(map[string]SignalReaderIdTree, len(inputType.Fields))}}

	i := 0
	for fieldName, _ := range inputType.Fields {
		com.Coms[i].Com = ICom{}
		com.InputMap.Children[fieldName] = SignalReaderIdTree{
			SignalReaderId: SignalReaderId{ReaderId: ReaderId{InternalP: true, InternalIdx: i}}}
		i++
	}

	com.Run(inputType, inputGetter, outputGetter, quit)
}
