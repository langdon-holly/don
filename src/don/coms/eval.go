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
		entry("prod", Prod()),
		entry("yet", Yet()),
	}), Gather()}),
	Macros: make(map[string]func(Context) func(EvalResult) EvalResult)}

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
			if name := param.Syntax(); name.Tag != NameSyntaxTag {
				panic("Non-name parameter to withoutField: " + name.String())
			} else if !name.LeftMarker && !name.RightMarker {
				return EvalResult{I(NullType.At(name.Name))}
			} else {
				panic("Marked parameter to withoutField: " + name.String())
			}
		}
	}
	ms["context"] = func(c Context) func(EvalResult) EvalResult {
		return func(param EvalResult) EvalResult {
			list := param.Syntax()
			if list.Tag != ListSyntaxTag {
				panic("Non-list parameter to context: " + list.String())
			} else if len(list.Children) > 0 {
				for _, listElem := range list.Children {
					if listElem.Tag != EmptyLineSyntaxTag {
						c.Com = Eval(listElem, c).Com()
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
				name := param0.Syntax()
				if name.Tag != NameSyntaxTag {
					panic("Non-name name parameter to def: " + name.String())
				} else if !name.LeftMarker && !name.RightMarker {
					return EvalResult{Pipe([]Com{
						Scatter(),
						Par([]Com{
							Pipe([]Com{c.Com.Copy(), I(NullType.At(name.Name))}),
							Pipe([]Com{Select(name.Name), param1.Com(), Deselect(name.Name)}),
						}),
						Gather(),
					})}
				} else {
					panic("Marked name parameter to def: " + name.String())
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
func eval(s Syntax, c Context) interface{} {
	switch s.Tag {
	case ListSyntaxTag:
		var factorComs []Com
		for _, factor := range s.Children {
			if factor.Tag != EmptyLineSyntaxTag {
				factorComs = append(factorComs, Eval(factor, c).Com())
			}
		}
		return Par(factorComs)
	case EmptyLineSyntaxTag:
		panic("Eval empty line")
	case ApplicationSyntaxTag:
		return Eval(s.Children[0], c).Apply(Eval(s.Children[1], c)).It
	case CompositionSyntaxTag:
		factorResults := make([]EvalResult, len(s.Children))
		for i, factor := range s.Children {
			factorResults[len(s.Children)-1-i] = Eval(factor, c)
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
	case NameSyntaxTag:
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
	case ISyntaxTag:
		return I(UnknownType)
	case QuotationSyntaxTag:
		return s.Children[0]
	}
	panic("Unreachable")
}

// c may be shared
func Eval(s Syntax, c Context) EvalResult {
	return EvalResult{eval(s, c)}
}
