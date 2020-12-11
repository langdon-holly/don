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
		parComs := make([]Com, len(s.Children))
		for i, line := range s.Children {
			parComs[i] = ComFromSyntax(line, context)
		}
		return ParCom(parComs)
	case syntax.SpacedSyntaxTag:
		pipeComs := make([]Com, len(s.Children))
		for i, subS := range s.Children {
			pipeComs[len(s.Children)-1-i] = ComFromSyntax(subS, context)
		}
		return PipeCom(pipeComs)
	case syntax.MCallSyntaxTag:
		if !s.LeftMarker && !s.RightMarker {
			child := s.Children[0]
			switch s.Name {
			case "rec":
				return RecCom{Inner: ComFromSyntax(child, context)}
			case "map":
				return MapCom{Com: ComFromSyntax(child, context)}
			case "~":
				return ComFromSyntax(child, context).Inverse()
			case "withoutField":
				if child.Tag != syntax.NameSyntaxTag {
					panic("Non-name parameter to withoutField")
				} else if !child.LeftMarker && !child.RightMarker {
					return ICom(NullType.At(child.Name))
				} else {
					panic("Marked parameter to withoutField: " + child.String())
				}
			case "bind":
				if child.Tag == syntax.ListSyntaxTag {
					for _, childChild := range child.Children {
						context = ComFromSyntax(childChild, context)
					}
					return context
				} else if panic("Non-list parameter to bind"); true {
				}
			}
			panic("Unknown macro: " + syntax.Syntax{Name: s.Name}.String())
		} else if panic("Marked macro name: " + syntax.Syntax{Name: s.Name}.String()); true {
		}
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
	case syntax.ContextSyntaxTag:
		return context
	case syntax.SandwichSyntaxTag:
		return PipeCom([]Com{
			ComFromSyntax(s.Children[0], context).Inverse(),
			ComFromSyntax(s.Children[1], context),
			ComFromSyntax(s.Children[0], context)})
	}
	panic("Unreachable")
}
