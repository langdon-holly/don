package coms

import "strconv"

import (
	. "don/core"
)

type ProdCom struct{}

func (ProdCom) Instantiate() ComInstance {
	return &prodInstance{inputType: StructType}
}

type prodInstance struct{ inputType, outputType DType }

func (pi *prodInstance) InputType() *DType  { return &pi.inputType }
func (pi *prodInstance) OutputType() *DType { return &pi.outputType }

func strToNat(s string) (nat int, err bool) {
	if s == "" {
		err = true
	} else if s[0] >= 58 {
		err = true
	} else if s[0] >= 49 {
		nat = int(s[0]) - 48
		for digit := range s[1:] {
			if digit < 48 || digit >= 58 {
				err = true
				return
			}
			nat *= 10
			nat += int(digit) - 48
		}
	} else if s == "0" {
		nat = 0
	} else {
		err = true
	}
	return
}

type pathLvl struct {
	NoUnit bool
	Fields map[string]struct{}
}

// len(Lvls) > 0
// After end of path: NullTypes if Positive, UnknownTypes otherwise
type path struct {
	Lvls          []pathLvl
	FinalPositive bool
	Unknowns      bool
}

func nullPath() path { return path{Lvls: []pathLvl{{NoUnit: true}}, FinalPositive: true} }

func (typesPath *path) MeetsLvl(ps pathLvl, psPositive bool, depth int) {
	if depth < len(typesPath.Lvls) {
		already := typesPath.Lvls[depth]
		already.NoUnit = already.NoUnit || ps.NoUnit
		alreadyPositive := depth < len(typesPath.Lvls)-1 || typesPath.FinalPositive
		if !psPositive {
		} else if alreadyPositive {
			for fieldName := range already.Fields {
				if _, inPsFields := ps.Fields[fieldName]; !inPsFields {
					delete(already.Fields, fieldName)
				}
			}
		} else if already.Fields = make(map[string]struct{}, len(ps.Fields)); true {
			for fieldName := range ps.Fields {
				already.Fields[fieldName] = struct{}{}
			}
			typesPath.FinalPositive = true
		}
		typesPath.Lvls[depth] = already
	} else if depth == len(typesPath.Lvls) && typesPath.FinalPositive && typesPath.Unknowns {
		typesPath.Lvls = append(typesPath.Lvls, ps)
		typesPath.FinalPositive = psPositive
	}
}
func (typesPath *path) JoinsLvl(ps pathLvl, psPositive bool, depth int) {
	if depth < len(typesPath.Lvls) {
		already := typesPath.Lvls[depth]
		already.NoUnit = already.NoUnit && ps.NoUnit
		alreadyPositive := depth < len(typesPath.Lvls)-1 || typesPath.FinalPositive
		if alreadyPositive && psPositive {
			if already.Fields == nil {
				already.Fields = make(map[string]struct{}, len(ps.Fields))
			}
			for fieldName := range ps.Fields {
				already.Fields[fieldName] = struct{}{}
			}
		} else if already.Fields = nil; true {
			typesPath.Lvls = typesPath.Lvls[:depth+1]
			typesPath.FinalPositive = false
			typesPath.Unknowns = true
		}
		typesPath.Lvls[depth] = already
	} else if typesPath.FinalPositive && !typesPath.Unknowns {
		for len(typesPath.Lvls) < depth {
			typesPath.Lvls = append(typesPath.Lvls, pathLvl{NoUnit: true})
		}
		typesPath.Lvls = append(typesPath.Lvls, ps)
		typesPath.FinalPositive = psPositive
		typesPath.Unknowns = !psPositive
	}
}

func typePath(t DType, depth int, typesPath *path) {
	lvl := pathLvl{NoUnit: t.NoUnit}
	if t.Positive {
		lvl.Fields = make(map[string]struct{}, len(t.Fields))
		for fieldName := range t.Fields {
			lvl.Fields[fieldName] = struct{}{}
		}
		typesPath.JoinsLvl(lvl, true, depth)
		for _, fieldType := range t.Fields {
			typePath(fieldType, depth+1, typesPath)
		}
	} else if typesPath.JoinsLvl(lvl, false, depth); true {
	}
}

func (typesPath path) Type() (t DType) {
	i := len(typesPath.Lvls) - 1
	if !typesPath.FinalPositive {
		t.NoUnit = typesPath.Lvls[i].NoUnit
		i--
	} else if typesPath.Unknowns {
		t = UnknownType
	} else if t = NullType; true {
	}
	for ; i >= 0; i-- {
		lvl := typesPath.Lvls[i]
		fields := make(map[string]DType, len(lvl.Fields))
		for fieldName := range lvl.Fields {
			fields[fieldName] = t
		}
		t = DType{NoUnit: lvl.NoUnit, Positive: true, Fields: fields}
	}
	return
}

func (pi *prodInstance) Types() (underdefined Error) {
	if pi.outputType.LTE(NullType) {
		pi.inputType = NullType
		return
	} else if !pi.inputType.Positive {
		return NewError("Negative input to prod")
	}
	pi.inputType.RemakeFields()

	var indexStrings []string
	for fieldName := range pi.inputType.Fields {
		if idx, badNat := strToNat(fieldName); !badNat {
			for idx >= len(indexStrings) {
				idxStr := strconv.Itoa(len(indexStrings))
				if _, exists := pi.inputType.Fields[idxStr]; exists {
					indexStrings = append(indexStrings, idxStr)
				} else if delete(pi.inputType.Fields, fieldName); true {
					break
				}
			}
		} else if delete(pi.inputType.Fields, fieldName); true {
		}
	}

	outputPath := nullPath()
	typePath(pi.outputType, 0, &outputPath)

	inputPaths := make([]path, len(pi.inputType.Fields))
	for i, idxStr := range indexStrings {
		inputPaths[i] = nullPath()
		typePath(pi.inputType.Fields[idxStr], 0, &inputPaths[i])
	}

	idxInOutPath := 0
	for j, inputPath := range inputPaths {
		for i, inputLvl := range inputPath.Lvls[:len(inputPath.Lvls)-1] {
			meetOutputLvl := inputLvl
			if i == 0 && j > 0 {
				meetOutputLvl.NoUnit = false
			}
			outputPath.MeetsLvl(meetOutputLvl, true, idxInOutPath)
			idxInOutPath++
			if !inputLvl.NoUnit {
				goto AFTER_INPUT_ITER
			}
		}
		if inputLvl := inputPath.Lvls[len(inputPath.Lvls)-1]; inputLvl.NoUnit {
			outputPath.MeetsLvl(inputLvl, inputPath.FinalPositive, idxInOutPath)
			goto AFTER_INPUT_ITER
		} else if !inputPath.FinalPositive || len(inputLvl.Fields) > 0 {
			goto AFTER_INPUT_ITER
		}
	}
	if len(inputPaths) > 0 {
		outputPath.MeetsLvl(pathLvl{}, true, idxInOutPath)
	} else if outputPath.MeetsLvl(pathLvl{NoUnit: true}, true, idxInOutPath); true {
	}
AFTER_INPUT_ITER:
	for i, idxStr := range indexStrings {
		inputFieldType := pi.inputType.Fields[idxStr]
		inputFieldType.Meets(inputPaths[i].Type())
		pi.inputType.Fields[idxStr] = inputFieldType
	}
	pi.outputType.Meets(outputPath.Type())

	return pi.inputType.Underdefined().Context("in input to prod")
}

func getFieldPath(pathChan chan<- []string, fieldPath []string, unitChan <-chan Unit) {
	for {
		<-unitChan
		pathChan <- fieldPath
	}
}

func getFieldPaths(pathChan chan<- []string, atPath []string, input Input) {
	if input.Unit != nil {
		fieldPath := make([]string, len(atPath))
		copy(fieldPath, atPath)
		go getFieldPath(pathChan, fieldPath, input.Unit)
	} else {
		subAtPath := append(atPath, "")
		for fieldName, subInput := range input.Fields {
			subAtPath[len(subAtPath)-1] = fieldName
			getFieldPaths(pathChan, subAtPath, subInput)
		}
	}
}

func (pi prodInstance) Run(input Input, output Output) {
	if len(pi.inputType.Fields) == 0 {
		return
	}
	var pathChans []<-chan []string
	for i := 0; i < len(pi.inputType.Fields); i++ {
		pathChan := make(chan []string)
		pathChans = append(pathChans, pathChan)

		fieldName := strconv.Itoa(i)
		subInput := input.Fields[fieldName]
		getFieldPaths(pathChan, nil, subInput)
	}
	for {
		currentOutput := output
		for _, pathChan := range pathChans {
			fieldPath := <-pathChan
			for _, fieldName := range fieldPath {
				currentOutput = currentOutput.Fields[fieldName]
			}
		}
		currentOutput.WriteUnit()
	}
}
