package core

type Com interface {
	Run(input Input, output Output, quit <-chan struct{})
	InputType() DType
	OutputType() DType
}
