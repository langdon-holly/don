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
	entry(">", coms.ScatterCom{}),
	entry("<", coms.GatherCom{}),
	entry("=>", coms.SplitCom{}),
	entry("<-", coms.MergeCom{}),
	entry("yet", coms.YetCom{}),
	entry("prod", coms.ProdCom{}),
	entry("unit", coms.ICom(UnitType)),
	entry("struct", coms.ICom(StructType)),
	entry("null", coms.NullCom{}),
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
		if s.LeftMarker {
			if s.RightMarker {
				panic("Doubly-marked macro")
			} else {
				return coms.PipeCom([]Com{coms.DeselectCom(s.Name), s.Child.ToCom(context), coms.SelectCom(s.Name)})
			}
		} else if s.RightMarker {
			return coms.PipeCom([]Com{coms.SelectCom(s.Name), s.Child.ToCom(context), coms.DeselectCom(s.Name)})
		} else {
			switch s.Name {
			case "rec":
				return coms.RecCom{Inner: s.Child.ToCom(context)}
			case "map":
				return coms.MapCom{Com: s.Child.ToCom(context)}
			}
			panic("Unknown macro")
		}
	case NameSyntaxTag:
		if s.LeftMarker {
			if s.RightMarker {
				panic("Doubly-marked macro: :" + s.Name + ":")
			} else {
				return coms.SelectCom(s.Name)
			}
		} else if s.RightMarker {
			return coms.DeselectCom(s.Name)
		} else {
			return coms.PipeCom([]Com{coms.DeselectCom(s.Name), context, coms.SelectCom(s.Name)})
		}
	}
	panic("Unreachable")
}
