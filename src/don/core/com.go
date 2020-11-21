package core

type Com interface {
	Instantiate() ComInstance
}

type ComInstance interface {
	InputType() *DType           /* Modified by Types(); may alias OutputType() */
	OutputType() *DType          /* Modified by Types(); may alias InputType() */
	Types() (underdefined Error) /* Mutates */
	Run(input Input, output Output)
}
