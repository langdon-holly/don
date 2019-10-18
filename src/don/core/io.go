package core

type Unit struct{}

type SyntaxTag int

const (
	StringSyntaxTag = SyntaxTag(iota)
	LolSyntaxTag
	MCallSyntaxTag
)

type String string

type Lol [][]Syntax

type MCall struct {
	Macro Syntax /* String or MCall */
	Arg   Syntax /* String or Lol */
}

type Syntax struct {
	Tag       SyntaxTag
	StringVal string
	LolVal    [][]Syntax
	MCallVal  *MCall
}

type StructIn map[string]Input
type StructOut map[string]Output

type Input struct {
	Unit   <-chan Unit
	Syntax <-chan Syntax
	Com <-chan Com
	Struct StructIn
}

type Output struct {
	Unit   chan<- Unit
	Syntax chan<- Syntax
	Com chan<- Com
	Struct StructOut
}
