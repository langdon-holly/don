package core

type Unit struct{}

type Ref struct {
	P     bool
	Input /* for P */
}

type StructIn map[string]Input
type StructOut map[string]Output

type Input struct {
	Unit   <-chan Unit
	Ref    <-chan Ref
	Com    <-chan Com
	Struct StructIn
}

type Output struct {
	Unit   chan<- Unit
	Ref    chan<- Ref
	Com    chan<- Com
	Struct StructOut
}
