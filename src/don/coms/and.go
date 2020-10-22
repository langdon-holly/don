package coms

import (
	. "don/core"
	//"don/types"
)

var And Com = PipeCom([]Com{
	ProdCom{},
	SplitMergeCom([]Com{
		PipeCom([]Com{SelectCom("T"), SelectCom("T"), DeselectCom("T")}),
		PipeCom([]Com{
			SplitMergeCom([]Com{
				PipeCom([]Com{SelectCom("T"), SelectCom("F")}),
				PipeCom([]Com{SelectCom("F"), SelectCom("T")}),
				PipeCom([]Com{SelectCom("F"), SelectCom("F")})}),
			DeselectCom("F")})})})

//type AndCom struct{}
//
//var andComInputType DType
//
//func init() {
//	andComInputType = MakeNStructType(2)
//	andComInputType.Fields["0"] = types.BoolType
//	andComInputType.Fields["1"] = types.BoolType
//}
//
//func (AndCom) Types(inputType, outputType *DType) (bad bool) {
//	*inputType, bad = MergeTypes(*inputType, andComInputType)
//	if bad {
//		return
//	}
//	*outputType, bad = MergeTypes(*outputType, types.BoolType)
//	return
//}
//func (AndCom) Type() (in, out DType) { return andComInputType, types.BoolType }
//func (AndCom) Run(inputType, outputType DType, inputGetter InputGetter, outputGetter OutputGetter) {
//	input := inputGetter.GetInput()
//	input0TChan := input.Fields["0"].Fields["T"].Unit
//	input0FChan := input.Fields["0"].Fields["F"].Unit
//	input1TChan := input.Fields["1"].Fields["T"].Unit
//	input1FChan := input.Fields["1"].Fields["F"].Unit
//
//	output := outputGetter.GetOutput()
//	outputTChan := output.Fields["T"].Unit
//	outputFChan := output.Fields["F"].Unit
//
//	for {
//		select {
//		case <-input0TChan:
//			select {
//			case <-input1TChan:
//				outputTChan <- Unit{}
//			case <-input1FChan:
//				outputFChan <- Unit{}
//			}
//		case <-input0FChan:
//			select {
//			case <-input1TChan:
//				outputFChan <- Unit{}
//			case <-input1FChan:
//				outputFChan <- Unit{}
//			}
//		}
//	}
//}
