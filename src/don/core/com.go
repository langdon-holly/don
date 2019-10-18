package core

type Com interface {
	OutputType(inputType PartialType) PartialType
	Run(inputType DType, input Input, output Output, quit <-chan struct{})
}
