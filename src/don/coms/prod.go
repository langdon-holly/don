package coms

import "strconv"

import (
	. "don/core"
)

type ProdCom struct{}

func strToNat(s string) (nat int, err bool) {
	if s == "" {
		err = true
		return
	}
	if s[0] < 49 {
		if s == "0" {
			return 0, false
		} else {
			err = true
			return
		}
	}
	if s[0] >= 58 {
		err = true
		return
	}
	nat = int(s[0]) - 48
	for digit := range s[1:] {
		if digit < 48 || digit >= 58 {
			err = true
			return
		}
		nat *= 10
		nat += int(digit) - 48
	}
	return
}

// TODO: No annihilation
func subtypeTop(t0, t1 DType) (merged DType, bad []string) {
	if t0.Tag == UnknownTypeTag {
		merged = t1
		return
	}
	if t1.Tag == UnknownTypeTag {
		merged = t0
		return
	}

	if t0.Tag != t1.Tag {
		bad = []string{"Cannot be both unit and struct"}
		return
	}

	if t0.Tag == UnitTypeTag {
		merged = UnitType
		return
	}

	// t0.Tag == t1.Tag == StructTypeTag
	merged = t0
	for fieldName := range t0.Fields {
		if _, inT1 := t1.Fields[fieldName]; !inT1 {
			bad = []string{"Different fields"}
			return
		}
	}
	if len(t0.Fields) < len(t1.Fields) {
		bad = []string{"Different fields"}
	}
	return
}

func mergeTops(fields map[string]struct{}, depth int, typesPath *[]map[string]struct{}, terminal bool) (bad []string) {
	if depth < len(*typesPath) {
		already := (*typesPath)[depth]
		for fieldName := range fields {
			if _, inAlready := already[fieldName]; !inAlready {
				bad = []string{"Different fields"}
			}
		}
		if len(fields) < len(already) {
			bad = []string{"Different fields"}
		}
	} else {
		if terminal {
			bad = []string{"Cannot be both unit and struct"}
			return
		}

		*typesPath = append(*typesPath, fields)
	}

	return
}

func terminatePath(depth int, typesPath []map[string]struct{}, terminal *bool) (bad []string) {
	if depth < len(typesPath) {
		bad = []string{"Cannot be both unit and struct"}
	}
	*terminal = true
	return
}

func typePath(t DType, depth int, typesPath *[]map[string]struct{}, terminal *bool) (bad []string) {
	if t.Tag == UnknownTypeTag {
		return
	}
	if t.Tag == UnitTypeTag {
		return terminatePath(depth, *typesPath, terminal)
	}

	// t.Tag == StructTypeTag

	fields := make(map[string]struct{}, len(t.Fields))
	for fieldName := range t.Fields {
		fields[fieldName] = struct{}{}
	}
	bad = mergeTops(fields, depth, typesPath, *terminal)
	if bad != nil {
		return
	}

	if t.Tag == StructTypeTag {
		for _, fieldType := range t.Fields {
			bad = typePath(fieldType, depth+1, typesPath, terminal)
			if bad != nil {
				return
			}
		}
	}

	return
}

func pathType(typesPath []map[string]struct{}, terminal bool) (t DType) {
	if terminal {
		t = UnitType
	}
	for i := len(typesPath) - 1; i >= 0; i-- {
		superT := MakeNStructType(len(typesPath[i]))
		for fieldName := range typesPath[i] {
			superT.Fields[fieldName] = t
		}
		t = superT
	}
	return
}

func (ProdCom) Types(inputType, outputType *DType) (bad []string, done bool) {
	if inputType.Tag == UnitTypeTag {
		bad = []string{"Unit prod input"}
		return
	}
	if inputType.Tag == UnknownTypeTag {
		return
	}

	var indexStrings []string
	for fieldName := range inputType.Fields {
		idx, badNat := strToNat(fieldName)
		if badNat {
			bad = []string{"Unnatural field name in prod input"}
			return
		}
		for len(indexStrings) <= idx {
			idxStr := strconv.FormatInt(int64(len(indexStrings)), 10)
			indexStrings = append(indexStrings, idxStr)
			if _, exists := inputType.Fields[idxStr]; !exists {
				bad = []string{"Input to prod skips field " + idxStr}
				return
			}
		}
	}

	var outputPath []map[string]struct{}
	terminalOutput := false
	bad = typePath(*outputType, 0, &outputPath, &terminalOutput)
	if bad != nil {
		bad = append(bad, "in bad prod output")
		return
	}

	inputPaths := make([][]map[string]struct{}, len(inputType.Fields))
	terminalInputs := make([]bool, len(inputType.Fields))
	for i, idxStr := range indexStrings {
		bad = typePath(inputType.Fields[idxStr], 0, &inputPaths[i], &terminalInputs[i])
		if bad != nil {
			bad = append(bad, "in bad prod input field "+idxStr)
			return
		}
	}

	idxInOutPath := 0
	terminalInput := true
	for j, inputPath := range inputPaths {
		for i, inputLevel := range inputPath {
			bad = mergeTops(inputLevel, idxInOutPath, &outputPath, terminalOutput)
			if bad != nil {
				bad = append(bad, "in prod matching output with input field "+indexStrings[i])
				return
			}
			idxInOutPath++
		}
		if !terminalInputs[j] {
			terminalInput = false
			break
		}
	}
	if terminalInput {
		terminatePath(idxInOutPath, outputPath, &terminalOutput)
		if bad != nil {
			bad = append(bad, "in prod matching output with input termination")
			return
		}
	}

	inputType.RemakeFields()
	for i, idxStr := range indexStrings {
		inputType.Fields[idxStr] = pathType(inputPaths[i], terminalInputs[i])
	}
	*outputType = pathType(outputPath, terminalOutput)

	done = inputType.Minimal()
	return
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

func (ProdCom) Run(inputType, outputType DType, input Input, output Output) {
	var pathChans []<-chan []string
	for i := 0; i < len(inputType.Fields); i++ {
		pathChan := make(chan []string)
		pathChans = append(pathChans, pathChan)

		fieldName := strconv.FormatInt(int64(i), 10)
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
