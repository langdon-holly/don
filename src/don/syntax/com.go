package syntax

import (
	"don/coms"
	. "don/core"
)

func entry(fieldName string, inner Com) Com {
	return coms.PipeCom([]Com{
		coms.SelectCom(fieldName),
		inner,
		coms.DeselectCom(fieldName)})
}

var DefContext = coms.PipeCom([]Com{coms.ScatterCom{}, coms.ParCom([]Com{
	entry("I", coms.ICom(UnknownType)),
	entry("unit", coms.ICom(UnitType)),
	entry("fields", coms.ICom(FieldsType)),
	entry("null", coms.NullCom{}),
	entry("<", coms.GatherCom{}),
	entry(">", coms.ScatterCom{}),
	entry("<|", coms.MergeCom{}),
	entry("|>", coms.ChooseCom{}),
	entry("<||", coms.JoinCom{}),
	entry("||>", coms.ForkCom{}),
	entry("prod", coms.ProdCom{}),
	entry("yet", coms.YetCom{}),
}), coms.GatherCom{}})

func (s Syntax) ToCom(context Com) Com {
	switch s.Tag {
	case ListSyntaxTag:
		parComs := make([]Com, len(s.Children))
		for i, line := range s.Children {
			parComs[i] = line.ToCom(context)
		}
		return coms.ParCom(parComs)
	case SpacedSyntaxTag:
		pipeComs := make([]Com, len(s.Children))
		for i, subS := range s.Children {
			pipeComs[len(s.Children)-1-i] = subS.ToCom(context)
		}
		return coms.PipeCom(pipeComs)
	case MCallSyntaxTag:
		if !s.LeftMarker && !s.RightMarker {
			child := s.Children[0]
			switch s.Name {
			case "rec":
				return coms.RecCom{Inner: child.ToCom(context)}
			case "map":
				return coms.MapCom{Com: child.ToCom(context)}
			case "~":
				return child.ToCom(context).Inverse()
			case "withoutField":
				if child.Tag != NameSyntaxTag {
					panic("Non-name parameter to withoutField")
				} else if !child.LeftMarker && !child.RightMarker {
					return coms.ICom(NullType.At(child.Name))
				} else if panic("Marked parameter to withoutField"); true {
				}
			}
			panic("Unknown macro")
		} else if panic("Marked macro name"); true {
		}
	case NameSyntaxTag:
		if s.LeftMarker {
			if s.RightMarker {
				panic("Doubly-marked variable: :" + s.Name + ":")
			} else {
				return coms.SelectCom(s.Name)
			}
		} else if s.RightMarker {
			return coms.DeselectCom(s.Name)
		} else {
			return coms.PipeCom([]Com{coms.DeselectCom(s.Name), context, coms.SelectCom(s.Name)})
		}
	case ContextSyntaxTag:
		return context
	case SandwichSyntaxTag:
		return coms.PipeCom([]Com{
			s.Children[0].ToCom(context).Inverse(),
			s.Children[1].ToCom(context),
			s.Children[0].ToCom(context)})
	}
	panic("Unreachable")
}
