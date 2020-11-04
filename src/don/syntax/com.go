package syntax

import "strconv"

import (
	"don/coms"
	. "don/core"
)

type Context struct {
	Bindings map[string]Com
	Parent   *Context
}

// May mutate c and its ancestors
func (c Context) Get(name string) (Com, bool) {
	if com, bound := c.Bindings[name]; bound {
		return com, true
	} else if c.Parent != nil {
		com, bound := c.Parent.Get(name)
		if bound {
			c.Bindings[name] = com
		}
		return com, bound
	}
	return nil, false
}

var DefContext = Context{Bindings: make(map[string]Com, 2)}

func init() {
	DefContext.Bindings["I"] = coms.ICom{}
	DefContext.Bindings["init"] = coms.InitCom{}
	DefContext.Bindings["split"] = coms.SplitCom{}
	DefContext.Bindings["merge"] = coms.MergeCom{}
	DefContext.Bindings["yet"] = coms.YetCom{}
	DefContext.Bindings["prod"] = coms.ProdCom{}
	DefContext.Bindings["unit"] = coms.UnitCom{}
	DefContext.Bindings["struct"] = coms.StructCom{}
	DefContext.Bindings["null"] = coms.NullCom{}
}

func (s Syntax) ToCom(context Context) Com {
	switch s.Tag {
	case ListSyntaxTag:
		splitMergeComs := make([]Com, len(s.Children))
		for i, line := range s.Children {
			subCom := line.ToCom(context)
			if s.LeftMarker {
				subCom = coms.PipeCom([]Com{subCom, coms.DeselectCom(strconv.FormatInt(int64(i), 10))})
			}
			if s.RightMarker {
				subCom = coms.PipeCom([]Com{coms.SelectCom(strconv.FormatInt(int64(i), 10)), subCom})
			}
			splitMergeComs[i] = subCom
		}
		return coms.SplitMergeCom(splitMergeComs)
	case SpacedSyntaxTag:
		pipeComs := make([]Com, len(s.Children))
		for i, subS := range s.Children {
			pipeComs[len(s.Children)-1-i] = subS.ToCom(context)
		}
		return coms.PipeCom(pipeComs)
	case MCallSyntaxTag:
		if s.LeftMarker && s.RightMarker {
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
				return coms.PipeCom([]Com{coms.SelectCom(s.Name), coms.DeselectCom(s.Name)})
			} else {
				return coms.SelectCom(s.Name)
			}
		} else if s.RightMarker {
			return coms.DeselectCom(s.Name)
		} else if val, bound := context.Get(s.Name); bound {
			return val
		}
		panic("Unknown macro: " + s.Name)
	}
	panic("Unreachable")
}
