package syntax

import "strconv"

//import "fmt"

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
	}
	if c.Parent != nil {
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
	DefContext.Bindings["and"] = coms.And
	DefContext.Bindings["prod"] = coms.ProdCom{}
	DefContext.Bindings["unit"] = coms.UnitCom{}
}

func (s Syntax) ToCom(context Context) Com {
	switch s.Tag {
	//case BindSyntaxTag:
	//	if len(s.Children) < 2 {
	//		if len(s.Children) < 1 {
	//			panic("Bind value syntax")
	//		}
	//		pipeComs := make([]Com, len(s.Children[0]))
	//		for i, subS := range s.Children[0] {
	//			pipeComs[len(s.Children[0])-1-i] = subS.ToCom(context)
	//		}
	//		return coms.PipeCom(pipeComs)
	//	} else {
	//		subcontext := Context{
	//			Bindings: make(map[string]Com, 1),
	//			Parent:   &context}
	//		binding := s.Children[len(s.Children)-1]
	//		if len(binding) != 2 || binding[0].Tag != MacroSyntaxTag {
	//			panic("Bind binding syntax")
	//		}
	//		subcontext.Bindings[binding[0].Name] = binding[1].ToCom(context)
	//		return Syntax{Tag: BindSyntaxTag, Children: s.Children[:len(s.Children)-1]}.ToCom(subcontext)
	//	}
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
		} else {
			val, bound := context.Get(s.Name)
			if !bound {
				panic("Unknown macro: " + s.Name)
			}
			return val
		}
	}
	panic("Unreachable")
}
