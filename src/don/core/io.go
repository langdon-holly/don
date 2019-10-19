package core

type Unit struct{}

type Ref struct {
	P     bool
	Input /* for P */
}

type Input struct {
	Unit   <-chan Unit
	Ref    <-chan Ref
	Struct map[string]Input
}

type Output struct {
	Unit   []chan<- Unit
	Ref    []chan<- Ref
	Struct map[string]Output
}

func (o Output) WriteUnit() {
	for _, oChan := range o.Unit {
		oChan <- Unit{}
	}
}

func (o Output) WriteRef(val Ref) {
	for _, oChan := range o.Ref {
		oChan <- val
	}
}
