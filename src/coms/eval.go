package coms

import (
	. "don/core"
	"don/syntax"
)

type EvalResult struct{ It interface{} }

func (r EvalResult) Com() Com              { return r.It.(Com) }
func (r EvalResult) Syntax() syntax.Syntax { return r.It.(syntax.Syntax) }
func (r EvalResult) Apply(param EvalResult) EvalResult {
	return EvalResult{r.It.(func(EvalResult) interface{})(param)}
}

type Context *struct {
	Entries map[string]interface{}
	Parent  Context
}

var DefContext = new(struct {
	Entries map[string]interface{}
	Parent  Context
})

func init() {
	DefContext.Entries = make(map[string]interface{})

	DefContext.Entries["unit"] = I(UnitType)
	DefContext.Entries["fields"] = I(FieldsType)

	DefContext.Entries["map"] = func(param EvalResult) interface{} {
		return Map(param.Com())
	}
	DefContext.Entries["~"] = func(param EvalResult) interface{} {
		return param.Com().Invert()
	}
	DefContext.Entries["withoutField"] = func(param EvalResult) interface{} {
		if named := param.Syntax().(syntax.Named); !named.LeftMarker && !named.RightMarker {
			return I(NullType.AtHigh(named.Name))
		} else {
			panic("Marked parameter to withoutField: " + named.String())
		}
	}
	DefContext.Entries["#"] = func(EvalResult) interface{} { return Null }
	DefContext.Entries["sandwich"] = func(param0 EvalResult) interface{} {
		return func(param1 EvalResult) interface{} {
			return Pipe([]Com{param0.Com().Copy().Invert(), param1.Com(), param0.Com()})
		}
	}
}

// c may be shared
func eval(ss syntax.Syntax, c Context) interface{} {
	switch s := ss.(type) {
	case syntax.Disjunction:
		var disjunctComs []Com
		for _, disjunct := range s.Disjuncts {
			if _, emptyp := disjunct.(syntax.EmptyLine); !emptyp {
				disjunctComs = append(disjunctComs, Eval(disjunct, c).Com())
			}
		}
		return Disjunction(disjunctComs)
	case syntax.Conjunction:
		var conjunctComs []Com
		for _, conjunct := range s.Conjuncts {
			if _, emptyp := conjunct.(syntax.EmptyLine); !emptyp {
				conjunctComs = append(conjunctComs, Eval(conjunct, c).Com())
			}
		}
		return Conjunction(conjunctComs)
	case syntax.EmptyLine:
		panic("Eval empty line")
	case syntax.Application:
		return Eval(s.Com, c).Apply(Eval(s.Arg, c)).It
	case syntax.Bind:
		named := s.Var.(syntax.Named)
		if named.LeftMarker || named.RightMarker {
			panic("Marked variable as bind variable: " + named.String())
		}
		return func(arg EvalResult) interface{} {
			subC := &struct {
				Entries map[string]interface{}
				Parent  Context
			}{Entries: make(map[string]interface{}, 1), Parent: c}
			subC.Entries[named.Name] = arg.It
			return Eval(s.Body, subC).It
		}
	case syntax.Composition:
		factorResults := make([]EvalResult, len(s.Factors))
		for i, factor := range s.Factors {
			factorResults[len(s.Factors)-1-i] = Eval(factor, c)
		}
		if _, macrosp := factorResults[0].It.(func(EvalResult) interface{}); macrosp {
			return func(param EvalResult) interface{} {
				for _, factorResult := range factorResults {
					param = factorResult.Apply(param)
				}
				return param
			}
		} else {
			factorComs := make([]Com, len(factorResults))
			for i, factorResult := range factorResults {
				factorComs[i] = factorResult.Com()
			}
			return Pipe(factorComs)
		}
	case syntax.Named:
		if s.LeftMarker {
			if s.RightMarker {
				panic("Doubly-marked variable: " + s.String())
			} else {
				return Select(s.Name)
			}
		} else if s.RightMarker {
			return Deselect(s.Name)
		} else {
			for ; c != nil; c = c.Parent {
				if entry, ok := c.Entries[s.Name]; !ok {
				} else if com, isCom := entry.(Com); isCom {
					return com.Copy()
				} else {
					return entry
				}
			}
			return Null
		}
	case syntax.ISyntax:
		return I(UnknownType)
	case syntax.Quote:
		return s.Syntax
	}
	panic("Unreachable")
}

// c may be shared
func Eval(s syntax.Syntax, c Context) EvalResult { return EvalResult{eval(s, c)} }
