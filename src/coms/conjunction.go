package coms

import . "don/core"

func Conjunction(conjuncts []Com) Com { return ConjunctionCom{Par(conjuncts)}.simplify() }

type ConjunctionCom struct{ par ParCom }

func (cc ConjunctionCom) simplify() Com {
	if len(cc.par.Inners) == 0 {
		return Null
	} else if len(cc.par.Inners) == 1 {
		for _, onlyInner := range cc.par.Inners {
			return onlyInner
		}
		panic("Unreachable")
	} else {
		return cc
	}
}

func (cc ConjunctionCom) InputType() DType  { return cc.par.InputType }
func (cc ConjunctionCom) OutputType() DType { return cc.par.OutputType }
func (cc ConjunctionCom) MeetTypes(inputType, outputType DType) Com {
	cc.par.MeetTypes(inputType, outputType)
	return cc.simplify()
}
func (cc ConjunctionCom) Underdefined() Error {
	return cc.par.Underdefined("conjunction")
}
func (cc ConjunctionCom) Copy() Com   { cc.par = cc.par.Copy(); return cc }
func (cc ConjunctionCom) Invert() Com { cc.par.Invert(); return cc }

func (cc ConjunctionCom) TypedCom(tcb TypedComBuilder /* mutated */, inputMap, outputMap TypeMap) {
	cc.par.TypedCom(
		tcb,
		inputMap,
		outputMap,
		func(inputVar Var, innerInputVars []Var) {
			for _, outputVar := range innerInputVars {
				tcb.Equate(inputVar, outputVar)
			}
		},
		func(outputVar Var, innerOutputVars []Var) {
			for _, inputVar := range innerOutputVars {
				tcb.Equate(outputVar, inputVar)
			}
		},
	)
}
