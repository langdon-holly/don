package core

type GenCom interface {
	OutputType(inputType PartialType) PartialType
	Com(inputType DType) Com
}
