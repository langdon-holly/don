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

type Context struct {
	Com    Com
	Macros map[string]func(Context) func(EvalResult) EvalResult
}

func entry(fieldName string, inner Com) Com {
	return PipeCom([]Com{
		SelectCom(fieldName),
		inner,
		DeselectCom(fieldName)})
}

var DefContext = Context{
	Com: PipeCom([]Com{ScatterCom{}, ParCom([]Com{
		entry("I", ICom(UnknownType)),
		entry("unit", ICom(UnitType)),
		entry("fields", ICom(FieldsType)),
		entry("null", NullCom{}),
		entry("<", GatherCom{}),
		entry(">", ScatterCom{}),
		entry("<|", MergeCom{}),
		entry("|>", ChooseCom{}),
		entry("<||", JoinCom{}),
		entry("||>", ForkCom{}),
		entry("prod", ProdCom{}),
		entry("yet", YetCom{}),
	}), GatherCom{}}),
	Macros: make(map[string]func(Context) func(EvalResult) EvalResult)}

func init() {
	ms := DefContext.Macros
	ms["rec"] = func(_ Context) func(EvalResult) EvalResult {
		return func(param EvalResult) EvalResult {
			return EvalResult{RecCom{Inner: param.Com()}}
		}
	}
	ms["map"] = func(_ Context) func(EvalResult) EvalResult {
		return func(param EvalResult) EvalResult {
			return EvalResult{MapCom{Com: param.Com()}}
		}
	}
	ms["~"] = func(_ Context) func(EvalResult) EvalResult {
		return func(param EvalResult) EvalResult {
			return EvalResult{param.Com().Inverse()}
		}
	}
	ms["withoutField"] = func(_ Context) func(EvalResult) EvalResult {
		return func(param EvalResult) EvalResult {
			if name := param.Syntax(); name.Tag != NameSyntaxTag {
				panic("Non-name parameter to withoutField: " + name.String())
			} else if !name.LeftMarker && !name.RightMarker {
				return EvalResult{ICom(NullType.At(name.Name))}
			} else {
				panic("Marked parameter to withoutField: " + name.String())
			}
		}
	}
	ms["context"] = func(c Context) func(EvalResult) EvalResult {
		return func(param EvalResult) EvalResult {
			list := param.Syntax()
			if list.Tag == ListSyntaxTag {
				for _, listElem := range list.Children {
					if listElem.Tag != EmptyLineSyntaxTag {
						c.Com = Eval(listElem, c).Com()
					}
				}
				return EvalResult{c.Com}
			} else {
				panic("Non-list parameter to context: " + list.String())
			}
		}
	}
	ms["#"] = func(_ Context) func(EvalResult) EvalResult {
		return func(_ EvalResult) EvalResult {
			return EvalResult{NullCom{}}
		}
	}
	ms["##"] = func(c Context) func(EvalResult) EvalResult {
		return func(_ EvalResult) EvalResult {
			return EvalResult{c.Com}
		}
	}
	ms["def"] = func(c Context) func(EvalResult) EvalResult {
		return func(param0 EvalResult) EvalResult {
			return EvalResult{func(param1 EvalResult) EvalResult {
				name := param0.Syntax()
				if name.Tag != NameSyntaxTag {
					panic("Non-name name parameter to def: " + name.String())
				} else if !name.LeftMarker && !name.RightMarker {
					return EvalResult{PipeCom([]Com{
						ScatterCom{},
						ParCom([]Com{
							PipeCom([]Com{c.Com, ICom(NullType.At(name.Name))}),
							PipeCom([]Com{
								SelectCom(name.Name),
								param1.Com(),
								DeselectCom(name.Name),
							}),
						}),
						GatherCom{},
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
				return EvalResult{PipeCom([]Com{
					param0.Com().Inverse(),
					param1.Com(),
					param0.Com()})}
			}}
		}
	}
}

func eval(s Syntax, c Context) interface{} {
	switch s.Tag {
	case ListSyntaxTag:
		var parComs []Com
		for _, line := range s.Children {
			if line.Tag != EmptyLineSyntaxTag {
				parComs = append(parComs, Eval(line, c).Com())
			}
		}
		return ParCom(parComs)
	case EmptyLineSyntaxTag:
		panic("Eval empty line")
	case ApplicationSyntaxTag:
		return Eval(s.Children[0], c).Apply(Eval(s.Children[1], c)).It
	case CompositionSyntaxTag:
		rs := make([]EvalResult, len(s.Children))
		for i, subS := range s.Children {
			rs[len(s.Children)-1-i] = Eval(subS, c)
		}
		if _, macrosp := rs[0].It.(func(EvalResult) EvalResult); macrosp {
			return func(param EvalResult) EvalResult {
				for _, subR := range rs {
					param = subR.Apply(param)
				}
				return param
			}
		} else {
			pipeComs := make([]Com, len(rs))
			for i, subR := range rs {
				pipeComs[i] = subR.Com()
			}
			return PipeCom(pipeComs)
		}
	case NameSyntaxTag:
		if s.LeftMarker {
			if s.RightMarker {
				panic("Doubly-marked variable: " + s.String())
			} else {
				return SelectCom(s.Name)
			}
		} else if s.RightMarker {
			return DeselectCom(s.Name)
		} else if macroEntry, ok := c.Macros[s.Name]; ok {
			return macroEntry(c)
		} else {
			return PipeCom([]Com{DeselectCom(s.Name), c.Com, SelectCom(s.Name)})
		}
	case QuotationSyntaxTag:
		return s.Children[0]
	}
	panic("Unreachable")
}
func Eval(s Syntax, c Context) EvalResult {
	return EvalResult{eval(s, c)}
}
