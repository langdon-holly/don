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
		subComs := make([]Com, len(s.Children))
		for i, line := range s.Children {
			subComs[i] = line.ToCom(context)
			if s.LeftMarker {
				indexStr := strconv.Itoa(i)
				subComs[i] = coms.PipeCom([]Com{subComs[i], coms.DeselectCom(indexStr)})
			}
			if s.RightMarker {
				indexStr := strconv.Itoa(i)
				subComs[i] = coms.PipeCom([]Com{coms.SelectCom(indexStr), subComs[i]})
			}
		}
		return coms.ParCom(subComs)
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
		} else if val, bound := context.Get(s.Name); bound {
			return val
		} else if panic("Unknown macro: " + s.Name); true {
		}
	}
	panic("Unreachable")
}
