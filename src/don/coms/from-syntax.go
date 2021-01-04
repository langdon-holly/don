package coms

import (
	. "don/core"
	. "don/syntax"
)

var DefMMContext = make(map[string]func(Syntax, Com) func(Syntax, Com) Com)

func init() {
	DefMMContext["def"] = func(param0 Syntax, context0 Com) func(Syntax, Com) Com {
		return func(param1 Syntax, context1 Com) Com {
			if param0.Tag != CompositionSyntaxTag ||
				len(param0.Children) != 1 ||
				param0.Children[0].Tag != NameSyntaxTag {
				panic("Non-name name parameter to def")
			} else if name := param0.Children[0]; !name.LeftMarker && !name.RightMarker {
				return PipeCom([]Com{
					ScatterCom{},
					ParCom([]Com{
						PipeCom([]Com{context0, ICom(NullType.At(name.Name))}),
						PipeCom([]Com{
							SelectCom(name.Name),
							ComFromSyntax(param1, context1),
							DeselectCom(name.Name),
						}),
					}),
					GatherCom{},
				})
			} else {
				panic("Marked name parameter to def")
			}
		}
	}
	DefMMContext["sandwich"] = func(param0 Syntax, context0 Com) func(Syntax, Com) Com {
		return func(param1 Syntax, context1 Com) Com {
			return PipeCom([]Com{
				ComFromSyntax(param0, context0).Inverse(),
				ComFromSyntax(param1, context1),
				ComFromSyntax(param0, context0)})
		}
	}
}

func MMacroFromSyntax(s Syntax, context Com) func(Syntax, Com) func(Syntax, Com) Com {
	switch s.Tag {
	case ListSyntaxTag:
		panic("Unimplemented: " + s.String())
	case EmptyLineSyntaxTag:
		panic("Macro from empty line")
	case ApplicationSyntaxTag:
		panic("Unimplemented: " + s.String())
	case CompositionSyntaxTag:
		return func(param Syntax, context1 Com) func(Syntax, Com) Com {
			for i := len(s.Children) - 1; i >= 0; i-- {
				param = Syntax{
					Tag:      ApplicationSyntaxTag,
					Children: []Syntax{s.Children[i], param}}
			}
			return MacroFromSyntax(param, context1)
		}
	case NameSyntaxTag:
		if s.LeftMarker || s.RightMarker {
			panic("Marked macro name: " + s.String())
		} else if macro, ok := DefMMContext[s.Name]; ok {
			return macro
		} else if panic("Unknown macro: " + s.String()); true {
		}
	}
	panic("Unreachable")
}

var DefMContext = make(map[string]func(Syntax, Com) Com)

func init() {
	DefMContext["rec"] = func(param Syntax, context Com) Com {
		return RecCom{Inner: ComFromSyntax(param, context)}
	}
	DefMContext["map"] = func(param Syntax, context Com) Com {
		return MapCom{Com: ComFromSyntax(param, context)}
	}
	DefMContext["~"] = func(param Syntax, context Com) Com {
		return ComFromSyntax(param, context).Inverse()
	}
	DefMContext["withoutField"] = func(param Syntax, context Com) Com {
		if param.Tag != NameSyntaxTag {
			panic("Non-name parameter to withoutField")
		} else if !param.LeftMarker && !param.RightMarker {
			return ICom(NullType.At(param.Name))
		} else {
			panic("Marked parameter to withoutField: " + param.String())
		}
	}
	DefMContext["context"] = func(param Syntax, context Com) Com {
		if param.Tag == ListSyntaxTag {
			for _, paramChild := range param.Children {
				if paramChild.Tag != EmptyLineSyntaxTag {
					context = ComFromSyntax(paramChild, context)
				}
			}
			return context
		} else {
			panic("Non-list parameter to context")
		}
	}
	DefMContext["#"] = func(param Syntax, context Com) Com {
		return NullCom{}
	}
	DefMContext["##"] = func(param Syntax, context Com) Com {
		return context
	}
}

func MacroFromSyntax(s Syntax, context Com) func(Syntax, Com) Com {
	switch s.Tag {
	case ListSyntaxTag:
		panic("Unimplemented: " + s.String())
	case EmptyLineSyntaxTag:
		panic("Macro from empty line")
	case ApplicationSyntaxTag:
		return MMacroFromSyntax(s.Children[0], context)(s.Children[1], context)
	case CompositionSyntaxTag:
		return func(param Syntax, context1 Com) Com {
			for i := len(s.Children) - 1; i >= 0; i-- {
				param = Syntax{
					Tag:      ApplicationSyntaxTag,
					Children: []Syntax{s.Children[i], param}}
			}
			return ComFromSyntax(param, context1)
		}
	case NameSyntaxTag:
		if s.LeftMarker || s.RightMarker {
			panic("Marked macro name: " + s.String())
		} else if macro, ok := DefMContext[s.Name]; ok {
			return macro
		} else if panic("Unknown macro: " + s.String()); true {
		}
	}
	panic("Unreachable")
}

func entry(fieldName string, inner Com) Com {
	return PipeCom([]Com{
		SelectCom(fieldName),
		inner,
		DeselectCom(fieldName)})
}

var DefContext = PipeCom([]Com{ScatterCom{}, ParCom([]Com{
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
}), GatherCom{}})

func ComFromSyntax(s Syntax, context Com) Com {
	switch s.Tag {
	case ListSyntaxTag:
		var parComs []Com
		for _, line := range s.Children {
			if line.Tag != EmptyLineSyntaxTag {
				parComs = append(parComs, ComFromSyntax(line, context))
			}
		}
		return ParCom(parComs)
	case EmptyLineSyntaxTag:
		panic("Com from empty line")
	case ApplicationSyntaxTag:
		return MacroFromSyntax(s.Children[0], context)(s.Children[1], context)
	case CompositionSyntaxTag:
		pipeComs := make([]Com, len(s.Children))
		for i, subS := range s.Children {
			pipeComs[len(s.Children)-1-i] = ComFromSyntax(subS, context)
		}
		return PipeCom(pipeComs)
	case NameSyntaxTag:
		if s.LeftMarker {
			if s.RightMarker {
				panic("Doubly-marked variable: " + s.String())
			} else {
				return SelectCom(s.Name)
			}
		} else if s.RightMarker {
			return DeselectCom(s.Name)
		} else {
			return PipeCom([]Com{DeselectCom(s.Name), context, SelectCom(s.Name)})
		}
	}
	panic("Unreachable")
}
