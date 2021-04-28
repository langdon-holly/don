package coms

import . "don/core"

func Disjunction(disjuncts []Com) Com { return DisjunctionCom{Par(disjuncts)}.simplify() }

type DisjunctionCom struct{ par ParCom }

func (dc DisjunctionCom) simplify() Com {
	if len(dc.par.Inners) == 0 {
		return Null
	} else if len(dc.par.Inners) == 1 {
		for _, onlyInner := range dc.par.Inners {
			return onlyInner
		}
		panic("Unreachable")
	} else {
		return dc
	}
}

func (dc DisjunctionCom) InputType() DType  { return dc.par.InputType }
func (dc DisjunctionCom) OutputType() DType { return dc.par.OutputType }
func (dc DisjunctionCom) MeetTypes(inputType, outputType DType) Com {
	dc.par.MeetTypes(inputType, outputType)
	return dc.simplify()
}
func (dc DisjunctionCom) Underdefined() Error {
	return dc.par.Underdefined("disjunction")
}
func (dc DisjunctionCom) Copy() Com   { dc.par = dc.par.Copy(); return dc }
func (dc DisjunctionCom) Invert() Com { dc.par.Invert(); return dc }

func (dc DisjunctionCom) TypedCom(tcb TypedComBuilder /* mutated */, inputMap, outputMap TypeMap) {
	dc.par.TypedCom(
		tcb,
		inputMap,
		outputMap,
		func(inputVar Var, innerInputVars []Var) {
			if len(innerInputVars) == 1 {
				tcb.Equate(inputVar, innerInputVars[0])
			} else {
				tcb.Add(&ChooseNode{In: inputVar, Out: innerInputVars})
			}
		},
		func(outputVar Var, innerOutputVars []Var) {
			if len(innerOutputVars) == 1 {
				tcb.Equate(outputVar, innerOutputVars[0])
			} else {
				tcb.Add(&MergeNode{In: innerOutputVars, Out: outputVar})
			}
		},
	)
}
