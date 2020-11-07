package core

type Com interface {
	Types(inputType, outputType *DType) (underdefined Error)
	Run(inputType, outputType DType, input Input, output Output)
}
