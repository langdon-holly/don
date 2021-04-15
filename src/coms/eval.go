package coms

import (
	. "don/core"
	. "don/syntax"
)

type EvalResult struct{ It interface{} }

func (r EvalResult) Com() Com       { return r.It.(Com) }
func (r EvalResult) Syntax() Syntax { return r.It.(Syntax) }
func (r EvalResult) Apply(param EvalResult) EvalResult {
	return r.It.(func(EvalResult) EvalResult)(param)
}

// Context arg to macro may be shared
type Context struct {
	Com    Com
	Macros map[string]func(Context) func(EvalResult) EvalResult
}

func entry(fieldName string, inner Com) Com {
	return Pipe([]Com{
		Select(fieldName),
		inner,
		Deselect(fieldName)})
}

var DefContext = Context{
	Com: Pipe([]Com{Scatter(), Par([]Com{
		entry("unit", I(UnitType)),
		entry("fields", I(FieldsType)),
		entry("<", Gather()),
		entry(">", Scatter()),
		entry("<|", Merge()),
		entry("|>", Choose()),
		entry("<||", Join()),
		entry("||>", Fork()),
	}), Gather()}),
	Macros: make(map[string]func(Context) func(EvalResult) EvalResult),
}

func init() {
	ms := DefContext.Macros
	ms["map"] = func(_ Context) func(EvalResult) EvalResult {
		return func(param EvalResult) EvalResult {
			return EvalResult{Map(param.Com())}
		}
	}
	ms["~"] = func(_ Context) func(EvalResult) EvalResult {
		return func(param EvalResult) EvalResult {
			return EvalResult{param.Com().Invert()}
		}
	}
	ms["withoutField"] = func(_ Context) func(EvalResult) EvalResult {
		return func(param EvalResult) EvalResult {
			if named := param.Syntax().(Named); !named.LeftMarker && !named.RightMarker {
				return EvalResult{I(NullType.At(named.Name))}
			} else {
				panic("Marked parameter to withoutField: " + named.String())
			}
		}
	}
	ms["context"] = func(c Context) func(EvalResult) EvalResult {
		return func(param EvalResult) EvalResult {
			if list := param.Syntax().(List); len(list.Factors) > 0 {
				for _, factor := range list.Factors {
					if _, emptyp := factor.(EmptyLine); !emptyp {
						c.Com = Eval(factor, c).Com()
					}
				}
				return EvalResult{c.Com}
			} else {
				return EvalResult{c.Com.Copy()}
			}
		}
	}
	ms["#"] = func(_ Context) func(EvalResult) EvalResult {
		return func(_ EvalResult) EvalResult {
			return EvalResult{Null}
		}
	}
	ms["##"] = func(c Context) func(EvalResult) EvalResult {
		return func(_ EvalResult) EvalResult {
			return EvalResult{c.Com.Copy()}
		}
	}
	ms["def"] = func(c Context) func(EvalResult) EvalResult {
		return func(param0 EvalResult) EvalResult {
			return EvalResult{func(param1 EvalResult) EvalResult {
				named := param0.Syntax().(Named)
				if !named.LeftMarker && !named.RightMarker {
					return EvalResult{Pipe([]Com{
						Scatter(),
						Par([]Com{
							Pipe([]Com{c.Com.Copy(), I(NullType.At(named.Name))}),
							Pipe([]Com{Select(named.Name), param1.Com(), Deselect(named.Name)}),
						}),
						Gather(),
					})}
				} else {
					panic("Marked named parameter to def: " + named.String())
				}
			}}
		}
	}
	ms["sandwich"] = func(_ Context) func(EvalResult) EvalResult {
		return func(param0 EvalResult) EvalResult {
			return EvalResult{func(param1 EvalResult) EvalResult {
				return EvalResult{Pipe([]Com{
					param0.Com().Copy().Invert(),
					param1.Com(),
					param0.Com()})}
			}}
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
	case Composition:
		factorResults := make([]EvalResult, len(s.Factors))
		for i, factor := range s.Factors {
			factorResults[len(s.Factors)-1-i] = Eval(factor, c)
		}
		if _, macrosp := factorResults[0].It.(func(EvalResult) EvalResult); macrosp {
			return func(param EvalResult) EvalResult {
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
		} else if macroEntry, ok := c.Macros[s.Name]; ok {
			return macroEntry(c)
		} else {
			return Pipe([]Com{Deselect(s.Name), c.Com.Copy(), Select(s.Name)})
		}
	case ISyntax:
		return I(UnknownType)
	case Quote:
		return s.Syntax
	}
	panic("Unreachable")
}

// c may be shared
func Eval(s Syntax, c Context) EvalResult {
	return EvalResult{eval(s, c)}
}
