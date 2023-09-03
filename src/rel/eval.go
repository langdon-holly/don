package rel

import (
	"fmt"
	"os"
	"path"
)

import (
	. "don/junctive"
	"don/syntax"
)

type Comment struct{}
type EvalResult struct{ It interface{} }

func (r EvalResult) Rel() Rel             { return r.It.(Rel) }
func (r EvalResult) Syntax() syntax.Words { return r.It.(syntax.Words) }
func (r EvalResult) Apply(param EvalResult) EvalResult {
	return EvalResult{r.It.(func(EvalResult) interface{})(param)}
}
func (r EvalResult) IsComment() bool    { _, isComment := r.It.(Comment); return isComment }
func (r EvalResult) List() []EvalResult { return r.It.([]EvalResult) }

type context *struct {
	Entries map[string]interface{}
	Parent  context
	Dir     string
}

var defContext = new(struct {
	Entries map[string]interface{}
	Parent  context
	Dir     string
})

func junction(junctive Junctive) func(EvalResult) interface{} {
	return func(param EvalResult) interface{} {
		junctReses := param.List()
		junctRels := make([]Rel, len(junctReses))
		for i, junctRes := range junctReses {
			junctRels[i] = junctRes.Rel()
		}
		return Junction(junctive, junctRels)
	}
}
func tupleJunction(junctive Junctive, leftTuple, rightTuple bool) func(EvalResult) interface{} {
	return func(param EvalResult) interface{} {
		junctReses := param.List()
		junctRels := make([]Rel, len(junctReses))
		for i, junctRes := range junctReses {
			var factorRels []Rel
			if leftTuple {
				factorRels = append(factorRels, Collect(junctive, fmt.Sprint(i)))
			}
			factorRels = append(factorRels, junctRes.Rel())
			if rightTuple {
				factorRels = append(factorRels, Select(junctive, fmt.Sprint(i)))
			}
			junctRels[i] = Composition(factorRels)
		}
		return Junction(junctive, junctRels)
	}
}

func init() {
	defContext.Entries = make(map[string]interface{})

	defContext.Entries["I"] = Composition(nil)
	defContext.Entries["false"] = Junction(ConJunctive, nil)
	defContext.Entries["true"] = Junction(DisJunctive, nil)
	defContext.Entries["~"] = func(param EvalResult) interface{} {
		return param.Rel().Convert()
	}
	defContext.Entries["!"] = func(param EvalResult) interface{} {
		if reses := param.List(); 0 < len(reses) {
			val := reses[0]
			for _, res := range reses[1:] {
				val = val.Apply(res)
			}
			return val.It
		} else {
			panic("Empty application")
		}
	}
	defContext.Entries[","] = junction(ConJunctive)
	defContext.Entries[",@"] = tupleJunction(ConJunctive, false, true)
	defContext.Entries["@,"] = tupleJunction(ConJunctive, true, false)
	defContext.Entries["@,@"] = tupleJunction(ConJunctive, true, true)
	defContext.Entries[";"] = junction(DisJunctive)
	defContext.Entries[";@"] = tupleJunction(DisJunctive, false, true)
	defContext.Entries["@;"] = tupleJunction(DisJunctive, true, false)
	defContext.Entries["@;@"] = tupleJunction(DisJunctive, true, true)
}

func pathJoin(dir, file string) string { return dir + "/" + file }

// c may be shared
func evalWord(w syntax.Word, c context) interface{} {
	if len(w.Specials) == 0 {
		name := w.Strings[0]
		for cNow := c; cNow != nil; cNow = cNow.Parent {
			if entry, ok := cNow.Entries[name]; !ok {
			} else if rel, isRel := entry.(Rel); isRel {
				return rel.Copy(make(map[*VarPtr]*VarPtr), make(map[*TypePtr]*TypePtr))
			} else {
				return entry
			}
		}

		dir := ""
		for cNow := c; dir == ""; cNow = cNow.Parent {
			if cNow == nil {
				panic("No filesystem")
			}
			dir = cNow.Dir
		}
		val := EvalFile(pathJoin(dir, name)).It
		if c.Entries == nil {
			c.Entries = make(map[string]interface{}, 1)
		}
		c.Entries[name] = val
		if rel, isRel := val.(Rel); isRel {
			return rel.Copy(make(map[*VarPtr]*VarPtr), make(map[*TypePtr]*TypePtr))
		} else {
			return val
		}
	} else {
		switch specialPayload := w.Specials[0].(type) {
		case syntax.WordSpecialDelimited:
			if len(w.Specials) > 1 {
				panic(
					"Overly special word: " + w.String(),
				)
			}
			if w.Strings[0] != "" || w.Strings[1] != "" {
				panic("Delimitation embedded in name: " + w.String())
			}
			if specialPayload.LeftDelim != specialPayload.RightDelim {
				panic(
					"Unmatched delimiters: " +
						specialPayload.LeftDelim.String() +
						" and " +
						specialPayload.RightDelim.String(),
				)
			}
			switch specialPayload.LeftDelim {
			case syntax.MaybeDelimNone:
				panic("Missing delimiters")
			case syntax.MaybeDelimParen:
				return evalWords(specialPayload.Words, c)
			case syntax.MaybeDelimBrace:
				return specialPayload.Words
			}
			panic("Unreachable")
		case syntax.WordSpecialJunct:
			if 1 < len(w.Specials) {
				panic(
					"Overly special word: " + w.String(),
				)
			}
			if w.Strings[0] == "" {
				if w.Strings[1] == "" {
					panic("Junct with no name: " + w.String())
				} else {
					return Collect(Junctive(specialPayload), w.Strings[1])
				}
			} else {
				if w.Strings[1] == "" {
					return Select(Junctive(specialPayload), w.Strings[0])
				} else {
					panic("Junct with two names: " + w.String())
				}
			}
		case syntax.WordSpecialCommentMarker:
			if w.Strings[0] != "" {
				panic("Named comment: " + w.String())
			}
			return Comment{}
		}
		panic("Unreachable")
	}
}

// c may be shared
func evalComposition(composition []syntax.Word /* non-nil */, c context) interface{} {
	var factorResults []EvalResult // No Comments
	for _, factor := range composition {
		if er := (EvalResult{evalWord(factor, c)}); !er.IsComment() {
			factorResults = append(factorResults, er)
		}
	}
	if len(factorResults) == 0 {
		panic("Empty composition")
	}
	if _, macrosp := factorResults[0].It.(func(EvalResult) interface{}); macrosp {
		return func(param EvalResult) interface{} {
			for i := len(factorResults) - 1; ; {
				param = factorResults[i].Apply(param)
				i--
				if i < 0 {
					break
				}
			}
			return param.It
		}
	} else {
		factorRels := make([]Rel, len(factorResults))
		for i, factorResult := range factorResults {
			factorRels[i] = factorResult.Rel()
		}
		return Composition(factorRels)
	}
}

// c may be shared
func EvalComposition(composition []syntax.Word /* non-nil */, c context) EvalResult {
	return EvalResult{evalComposition(composition, c)}
}

// c may be shared
func evalWords(ws syntax.Words, c context) interface{} {
	if 0 < len(ws.Compositions) {
		compositions := ws.Compositions[1:]
		reses := make([]EvalResult, len(compositions))
		for i, composition := range compositions {
			// 0 < len(composition) (by def.)
			reses[i] = EvalComposition(composition, c)
		}
		// 0 < len(ws.Compositions[0]) (by def.)
		return EvalComposition(ws.Compositions[0], c).Apply(EvalResult{reses}).It
	} else {
		panic("No words")
	}
}

func EvalFile(srcPath string) EvalResult {
	if file, err := os.Open(srcPath); err == nil {
		for {
			if dest, err := os.Readlink(srcPath); err != nil {
				break
			} else if dir, _ := path.Split(srcPath); true {
				srcPath = pathJoin(dir, dest)
			}
		}
		dir, _ := path.Split(srcPath)
		if dir == "" {
			dir = "."
		}

		return EvalResult{evalWords(syntax.Parse(file), &struct {
			Entries map[string]interface{}
			Parent  context
			Dir     string
		}{Entries: nil, Parent: defContext, Dir: dir})}
	} else {
		panic(err)
	}
}
