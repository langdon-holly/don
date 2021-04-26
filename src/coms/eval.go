package coms

import (
	. "don/core"
	. "don/syntax"
)

type EvalResult struct{ It interface{} }

func (r EvalResult) Com() Com       { return r.It.(Com) }
func (r EvalResult) Syntax() Syntax { return r.It.(Syntax) }
func (r EvalResult) Apply(param EvalResult) EvalResult {
	return EvalResult{r.It.(func(EvalResult) interface{})(param)}
}

type Context *struct {
	Entries map[string]interface{}
	Parent  Context
}

func entry(fieldName string, inner Com) Com {
	return Pipe([]Com{
		Select(fieldName),
		inner,
		Deselect(fieldName)})
}

var DefContext = new(struct {
	Entries map[string]interface{}
	Parent  Context
})

func init() {
	DefContext.Entries = make(map[string]interface{})

	DefContext.Entries["unit"] = I(UnitType)
	DefContext.Entries["fields"] = I(FieldsType)
	DefContext.Entries["<"] = Gather()
	DefContext.Entries[">"] = Scatter()
	DefContext.Entries["<|"] = Merge()
	DefContext.Entries["|>"] = Choose()
	DefContext.Entries["<||"] = Join()
	DefContext.Entries["||>"] = Fork()

	DefContext.Entries["map"] = func(param EvalResult) interface{} {
		return Map(param.Com())
	}
	DefContext.Entries["~"] = func(param EvalResult) interface{} {
		return param.Com().Invert()
	}
	DefContext.Entries["withoutField"] = func(param EvalResult) interface{} {
		if named := param.Syntax().(Named); !named.LeftMarker && !named.RightMarker {
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
func eval(ss Syntax, c Context) interface{} {
	switch s := ss.(type) {
	case List:
		var factorComs []Com
		for _, factor := range s.Factors {
			if _, emptyp := factor.(EmptyLine); !emptyp {
				factorComs = append(factorComs, Eval(factor, c).Com())
			}
		}
		return Par(factorComs)
	case EmptyLine:
		panic("Eval empty line")
	case Application:
		return Eval(s.Com, c).Apply(Eval(s.Arg, c)).It
	case Bind:
		named := s.Var.(Named)
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
	case Composition:
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
	case Named:
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
	case ISyntax:
		return I(UnknownType)
	case Quote:
		return s.Syntax
	}
	panic("Unreachable")
}

// c may be shared
func Eval(s Syntax, c Context) EvalResult { return EvalResult{eval(s, c)} }
