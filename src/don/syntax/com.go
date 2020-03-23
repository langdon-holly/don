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
	DefContext.Bindings["merge"] = coms.MergeCom{}
	DefContext.Bindings["prod"] = coms.ProdCom{}
}

func (s Syntax) ToCom(context Context) Com {
	switch s.Tag {
	case BindSyntaxTag:
		if len(s.Children) < 1 || len(s.Children[0]) != 1 {
			panic("Bind value syntax")
		}
		subcontext := Context{
			Bindings: make(map[string]Com),
			Parent:   &context}
		for i := len(s.Children) - 1; i >= 1; i-- {
			binding := s.Children[i]
			if len(binding) != 2 || binding[0].Tag != MacroSyntaxTag {
				panic("Bind binding syntax")
			}
			subcontext.Bindings[binding[0].Name] = binding[1].ToCom(context)
		}
		return s.Children[0][0].ToCom(subcontext)
	case BlockSyntaxTag:
		var leftAt, rightAt int
		if s.LeftAt {
			leftAt = 1
		}
		if s.RightAt {
			rightAt = 1
		}

		pipes := make([]Com, len(s.Children))
		for i, line := range s.Children {
			pipeComs := make([]Com, len(line)+leftAt+rightAt)
			if s.LeftAt {
				pipeComs[rightAt+len(line)] = coms.Deselect(strconv.FormatInt(int64(i), 10))
			}
			for j, subS := range line {
				pipeComs[rightAt+len(line)-1-j] = subS.ToCom(context)
			}
			if s.RightAt {
				pipeComs[0] = coms.SelectCom(strconv.FormatInt(int64(i), 10))
			}
			pipes[i] = coms.Pipe(pipeComs)
		}

		return coms.SplitMerge(pipes)
	case MCallSyntaxTag:
		switch s.Name {
		case "com":
			if len(s.Children) < 1 {
				panic("Empty [com] body")
			}

			pipes := make([]Com, len(s.Children))
			for i, line := range s.Children {
				pipeComs := make([]Com, len(line))
				for j, subS := range line {
					pipeComs[len(line)-1-j] = subS.ToCom(context)
				}
				pipes[i] = coms.Pipe(pipeComs)
			}
			return coms.ComCom(pipes)
		}
		panic("Unknown macro")
	case MacroSyntaxTag:
		val, bound := context.Get(s.Name)
		if !bound {
			panic("Unknown macro")
		}
		return val
	case SelectSyntaxTag:
		return coms.SelectCom(s.Name)
	case DeselectSyntaxTag:
		return coms.Deselect(s.Name)
	}
	panic("Unreachable")
}
