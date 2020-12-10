package core

type Com interface {
	Instantiate() ComInstance
	Inverse() Com
}

type ComInstance interface {
	InputType() *DType   /* Invalidated by Types; may alias OutputType(); after, call Types before Underdefined */
	OutputType() *DType  /* Invalidated by Types; may alias InputType(); after, call Types before Underdefined */
	Types()              /* Mutates */
	Underdefined() Error /* First, call Types */
	Run(input Input, output Output)
}
