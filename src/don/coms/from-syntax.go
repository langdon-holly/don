package coms

import (
	. "don/core"
	"don/syntax"
)

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

func ComFromSyntax(s syntax.Syntax, context Com) Com {
	switch s.Tag {
	case syntax.ListSyntaxTag:
		var parComs []Com
		for _, line := range s.Children {
			if line.Tag != syntax.EmptyLineSyntaxTag {
				parComs = append(parComs, ComFromSyntax(line, context))
			}
		}
		return ParCom(parComs)
	case syntax.EmptyLineSyntaxTag:
		panic("Com from empty line")
	case syntax.SpacedSyntaxTag:
		pipeComs := make([]Com, len(s.Children))
		for i, subS := range s.Children {
			pipeComs[len(s.Children)-1-i] = ComFromSyntax(subS, context)
		}
		return PipeCom(pipeComs)
	case syntax.MCallSyntaxTag:
		name := s.Children[0]
		if name.Tag != syntax.NameSyntaxTag {
			panic("Non-name macro name")
		} else if !name.LeftMarker && !name.RightMarker {
			param := s.Children[1]
			switch name.Name {
			case "rec":
				return RecCom{Inner: ComFromSyntax(param, context)}
			case "map":
				return MapCom{Com: ComFromSyntax(param, context)}
			case "~":
				return ComFromSyntax(param, context).Inverse()
			case "withoutField":
				if param.Tag != syntax.NameSyntaxTag {
					panic("Non-name parameter to withoutField")
				} else if !param.LeftMarker && !param.RightMarker {
					return ICom(NullType.At(param.Name))
				} else {
					panic("Marked parameter to withoutField: " + param.String())
				}
			case "context":
				if param.Tag == syntax.ListSyntaxTag {
					for _, paramChild := range param.Children {
						if paramChild.Tag != syntax.EmptyLineSyntaxTag {
							context = ComFromSyntax(paramChild, context)
						}
					}
					return context
				} else if panic("Non-list parameter to bind"); true {
				}
			case "#":
				return NullCom{}
			case "##":
				return context
			}
			panic("Unknown macro: " + name.String())
		} else if panic("Marked macro name: " + name.String()); true {
		}
	case syntax.SandwichSyntaxTag:
		return PipeCom([]Com{
			ComFromSyntax(s.Children[0], context).Inverse(),
			ComFromSyntax(s.Children[1], context),
			ComFromSyntax(s.Children[0], context)})
	case syntax.NameSyntaxTag:
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
