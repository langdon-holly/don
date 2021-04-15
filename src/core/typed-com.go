package core

type Var *struct{ _ bool }

func GenVar() Var { return &struct{ _ bool }{} }

type IOVar struct {
	Output chan<- struct{}
	Input  <-chan struct{}
}
type IO map[Var]IOVar

type TypedComNode interface {
	Substitute(subs map[Var]Var)         /* Mutates */
	Vars(vars /* mutated */ map[Var]int) /* The int is the number of writers */
	Run(io IO)                           /* Call with go */
}

func substituteVar(subs map[Var]Var, varA *Var) {
	if newVar, exists := subs[*varA]; exists {
		*varA = newVar
	}
}

type ChooseNode struct {
	In  Var
	Out []Var
}

// Mutates
func (cc *ChooseNode) Substitute(subs map[Var]Var) {
	substituteVar(subs, &cc.In)
	for i := range cc.Out {
		substituteVar(subs, &cc.Out[i])
	}
}

func (ChooseNode) Vars(map[Var]int) { panic("Unimplemented") }
func (ChooseNode) Run(IO)           { panic("Unimplemented") }

type MergeNode struct {
	In  []Var
	Out Var
}

// Mutates
func (mc *MergeNode) Substitute(subs map[Var]Var) {
	for i := range mc.In {
		substituteVar(subs, &mc.In[i])
	}
	substituteVar(subs, &mc.Out)
}

func (mc MergeNode) Vars(vars /* mutated */ map[Var]int) {
	for _, inVar := range mc.In {
		vars[inVar] = vars[inVar]
	}
	vars[mc.Out]++
}

func pipeUnit(outChan chan<- struct{}, inChan <-chan struct{}) {
	outChan <- <-inChan
}

func (mc MergeNode) Run(io IO) {
	outChan := io[mc.Out].Output
	for _, inVar := range mc.In {
		go pipeUnit(outChan, io[inVar].Input)
	}
}

// jth var in ith factor is In[i][j]
//
// OutStrides has entry per factor
// Given a jth var A[i] of each factor i, the corresponding output var is
// 	Out[A[0]*1 + ... + A[i]*OutStrides[i] + ...]
// OutStrides[0] == 1, if exists
// OutStrides[i+1] == OutStrides[i] * len(In[i])
// Logically, OutStrides[len(OutStrides)] == len(Out)
type ProdNode struct {
	In         [][]Var
	Out        []Var
	OutStrides []int
}

// Mutates
func (pc *ProdNode) Substitute(subs map[Var]Var) {
	for _, factor := range pc.In {
		for j := range factor {
			substituteVar(subs, &factor[j])
		}
	}
	for i := range pc.Out {
		substituteVar(subs, &pc.Out[i])
	}
}

func (pc ProdNode) Vars(vars /* mutated */ map[Var]int) {
	for _, factorVars := range pc.In {
		for _, inVar := range factorVars {
			vars[inVar] = vars[inVar]
		}
	}
	for _, outVar := range pc.Out {
		vars[outVar]++
	}
}

func notifyIndex(indexChan chan<- int, inChan <-chan struct{}, index int) {
	<-inChan
	indexChan <- index
}

func (pc ProdNode) Run(io IO) {
	outIdx := 0
	for factor, factorVars := range pc.In {
		indexChan := make(chan int)
		for index, factorVar := range factorVars {
			go notifyIndex(indexChan, io[factorVar].Input, index)
		}
		outIdx += <-indexChan * pc.OutStrides[factor]
	}
	io[pc.Out[outIdx]].Output <- struct{}{}
}

type TypeMap struct {
	Unit   Var
	Fields map[string]TypeMap
}

func MakeTypeMap(t DType) (tm TypeMap) {
	if !t.NoUnit {
		tm.Unit = GenVar()
	}
	tm.Fields = make(map[string]TypeMap, len(t.Fields))
	for fieldName, fieldType := range t.Fields {
		tm.Fields[fieldName] = MakeTypeMap(fieldType)
	}
	return
}

func (one TypeMap) forEachWith(many []TypeMap, fn func(one Var, many []Var)) {
	for fieldName, subOne := range one.Fields {
		var subMany []TypeMap
		for _, manyElem := range many {
			if manyElemField, ok := manyElem.Fields[fieldName]; ok {
				subMany = append(subMany, manyElemField)
			}
		}
		subOne.forEachWith(subMany, fn)
	}
	if one.Unit != nil {
		var manyVars []Var
		for _, manyElem := range many {
			if manyElem.Unit != nil {
				manyVars = append(manyVars, manyElem.Unit)
			}
		}
		fn(one.Unit, manyVars)
	}
}
func (one TypeMap) ForEachWith(many TypeMap, fn func(one Var, many []Var)) {
	fields := make([]TypeMap, len(many.Fields))
	{
		i := 0
		for _, field := range many.Fields {
			fields[i] = field
			i++
		}
	}
	one.forEachWith(fields, fn)
}

// Mutates
func (tm *TypeMap) Substitute(subs map[Var]Var) {
	if subs[tm.Unit] != nil {
		tm.Unit = subs[tm.Unit]
	}
	fields := make(map[string]TypeMap, len(tm.Fields))
	for fieldName, subTypeMap := range tm.Fields {
		subTypeMap.Substitute(subs)
		fields[fieldName] = subTypeMap
	}
	tm.Fields = fields
}

type TypedCom struct {
	Nodes               map[TypedComNode]struct{}
	InputMap, OutputMap TypeMap
}

type TypedComBuilder struct {
	Nodes map[TypedComNode]struct{}
	Eqs   map[Var]map[Var]struct{}
}

// Mutates
func (tcb TypedComBuilder) Add(node TypedComNode) { tcb.Nodes[node] = struct{}{} }
func (tcb TypedComBuilder) Equate(v0, v1 Var) {
	if tcb.Eqs[v0] == nil {
		tcb.Eqs[v0] = make(map[Var]struct{})
	}
	tcb.Eqs[v0][v1] = struct{}{}
	if tcb.Eqs[v1] == nil {
		tcb.Eqs[v1] = make(map[Var]struct{})
	}
	tcb.Eqs[v1][v0] = struct{}{}
}

func equivalenceClass(
	subs map[Var]Var, /* mutated */
	eqs map[Var]map[Var]struct{}, /* mutated */
	newVar, currVar Var,
) {
	subs[currVar] = newVar
	if nextVars, exists := eqs[currVar]; exists {
		delete(eqs, currVar)
		for nextVar := range nextVars {
			equivalenceClass(subs, eqs, newVar, nextVar)
		}
	}
}

func MakeTypedCom(com Com) (tc TypedCom) {
	var tcb TypedComBuilder
	tcb.Nodes = make(map[TypedComNode]struct{})
	tcb.Eqs = make(map[Var]map[Var]struct{})

	tc.Nodes = tcb.Nodes
	tc.InputMap = MakeTypeMap(com.InputType())
	tc.OutputMap = MakeTypeMap(com.OutputType())
	com.TypedCom(tcb, tc.InputMap, tc.OutputMap)

	subs := make(map[Var]Var)
	for len(tcb.Eqs) > 0 {
		var newVar Var
		for newVar = range tcb.Eqs {
			break
		}
		equivalenceClass(subs, tcb.Eqs, newVar, newVar)
	}
	for node := range tc.Nodes {
		node.Substitute(subs)
	}
	tc.InputMap.Substitute(subs)
	tc.OutputMap.Substitute(subs)

	return
}

func flattenChoose(chooses, choices map[Var]map[Var]struct{} /* mutated */, choose Var, uses map[Var]int) (root Var) {
	if uses[choose] == 2 && choices[choose] != nil {
		children := chooses[choose]
		var parent Var
		for parent = range choices[choose] {
			break
		}

		// delete choose
		delete(chooses, choose)
		delete(choices, choose)
		delete(chooses[parent], choose)
		for child := range children {
			delete(choices[child], choose)
		}

		root = flattenChoose(chooses, choices, parent, uses)
		for child := range children {
			chooses[root][child] = struct{}{}
			choices[child][root] = struct{}{}
		}
	} else {
		root = choose
	}
	return
}

func collectProdParts(
	choosesOrChoicesForProd /* mutated */ map[Var]map[Var]struct{},
	choicesOrChoosesForProd /* mutated */ map[Var]map[Var]struct{},
	choosesOrChoices /* mutated */ map[Var]map[Var]struct{},
	choicesOrChooses /* mutated */ map[Var]map[Var]struct{},
	chooseOrChoiceVar Var,
) {
	if chooseOrChoice, exists := choosesOrChoices[chooseOrChoiceVar]; exists {
		delete(choosesOrChoices, chooseOrChoiceVar)
		choosesOrChoicesForProd[chooseOrChoiceVar] = chooseOrChoice
		for choiceOrChooseVar := range chooseOrChoice {
			collectProdParts(
				choicesOrChoosesForProd,
				choosesOrChoicesForProd,
				choicesOrChooses,
				choosesOrChoices,
				choiceOrChooseVar,
			)
		}
	}
}

// Mutates
func (tc TypedCom) Determinate() {
	chooses := make(map[Var]map[Var]struct{})
	choices := make(map[Var]map[Var]struct{})
	uses := make(map[Var]int)
	for node := range tc.Nodes {
		switch n := node.(type) {
		case *ChooseNode:
			delete(tc.Nodes, node)
			chooses[n.In] = make(map[Var]struct{}, len(n.Out))
			uses[n.In]++
			for _, choiceVar := range n.Out {
				chooses[n.In][choiceVar] = struct{}{}
				if choices[choiceVar] == nil {
					choices[choiceVar] = make(map[Var]struct{})
				}
				choices[choiceVar][n.In] = struct{}{}
				uses[choiceVar]++
			}
		case *MergeNode:
			for _, inVar := range n.In {
				uses[inVar]++
			}
			uses[n.Out]++
		case *ProdNode:
			for _, factor := range n.In {
				for _, inVar := range factor {
					uses[inVar]++
				}
			}
			for _, outVar := range n.Out {
				uses[outVar]++
			}
		default:
			panic("Unreachable")
		}
	}
	for choose := range chooses {
		flattenChoose(chooses, choices, choose, uses)
	}
	for len(choices) > 0 {
		choosesForProd := make(map[Var]map[Var]struct{})
		choicesForProd := make(map[Var]map[Var]struct{})
		for choiceVar := range choices {
			collectProdParts(choicesForProd, choosesForProd, choices, chooses, choiceVar)
			break
		}

		var prod ProdNode
		prod.Out = make([]Var, len(choicesForProd))
		var choice0 map[Var]struct{}
		{
			var choiceVar0 Var
			for choiceVar0, choice0 = range choicesForProd {
				delete(choicesForProd, choiceVar0)
				break
			}
			prod.Out[0] = choiceVar0
		}
		prod.In = make([][]Var, len(choice0))
		chooseIdxs := make(map[Var]struct{ Factor, IdxInFactor int }, len(choosesForProd))
		{
			factor := 0
			for choice0Elem := range choice0 {
				chooseIdxs[choice0Elem] = struct{ Factor, IdxInFactor int }{factor, 0}
				prod.In[factor] = []Var{choice0Elem}
				factor++
			}
		}
		for _, choice := range choicesForProd {
			nOff := 0
			var inChoice0 Var
			for choice0Elem := range choice0 {
				if _, exists := choice[choice0Elem]; !exists {
					nOff++
					inChoice0 = choice0Elem
				}
			}
			if nOff == 1 {
				factor := chooseIdxs[inChoice0].Factor
				for choiceElem := range choice {
					if _, exists := choice0[choiceElem]; !exists {
						chooseIdxs[choiceElem] =
							struct{ Factor, IdxInFactor int }{factor, len(prod.In[factor])}
						prod.In[factor] = append(prod.In[factor], choiceElem)
						break
					}
				}
			}
		}
		prod.OutStrides = make([]int, len(choice0))
		{
			outStride := 1
			for i, factor := range prod.In {
				prod.OutStrides[i] = outStride
				outStride *= len(factor)
			}
		}
		for choiceVar, choice := range choicesForProd {
			outIdx := 0
			for choiceElem := range choice {
				chooseIdx := chooseIdxs[choiceElem]
				outIdx += chooseIdx.IdxInFactor * prod.OutStrides[chooseIdx.Factor]
			}
			prod.Out[outIdx] = choiceVar
		}
		tc.Nodes[&prod] = struct{}{}
	}
	if len(chooses) > 0 {
		panic("Uh oh")
	}
	return
}

type WriteMap struct {
	Unit   chan<- struct{}
	Fields map[string]WriteMap
}
type ReadMap struct {
	Unit   <-chan struct{}
	Fields map[string]ReadMap
}

func MakeWriteMap(io IO, inputMap TypeMap) (wMap WriteMap) {
	if inputMap.Unit != nil {
		wMap.Unit = io[inputMap.Unit].Output
	}
	wMap.Fields = make(map[string]WriteMap)
	for fieldName, subInputMap := range inputMap.Fields {
		wMap.Fields[fieldName] = MakeWriteMap(io, subInputMap)
	}
	return
}
func MakeReadMap(io IO, outputMap TypeMap) (rMap ReadMap) {
	if outputMap.Unit != nil {
		rMap.Unit = io[outputMap.Unit].Input
	}
	rMap.Fields = make(map[string]ReadMap)
	for fieldName, subOutputMap := range outputMap.Fields {
		rMap.Fields[fieldName] = MakeReadMap(io, subOutputMap)
	}
	return
}

func inputMapVars(vars /* mutated */ map[Var]int, inputMap TypeMap) {
	if inputMap.Unit != nil {
		vars[inputMap.Unit]++
	}
	for _, subInputMap := range inputMap.Fields {
		inputMapVars(vars, subInputMap)
	}
}

func runIOVar(output <-chan struct{}, input chan<- struct{}, nWriters int) {
	for ; nWriters > 0; nWriters-- {
		<-output
	}
	close(input)
}

func (tc TypedCom) Run() (wMap WriteMap, rMap ReadMap) {
	vars := make(map[Var]int)
	for node := range tc.Nodes {
		node.Vars(vars)
	}
	inputMapVars(vars, tc.InputMap)
	io := make(IO)
	for aVar, nWriters := range vars {
		output, input := make(chan struct{}), make(chan struct{})
		go runIOVar(output, input, nWriters)
		io[aVar] = IOVar{Output: output, Input: input}
	}
	for node := range tc.Nodes {
		go node.Run(io)
	}
	return MakeWriteMap(io, tc.InputMap), MakeReadMap(io, tc.OutputMap)
}
