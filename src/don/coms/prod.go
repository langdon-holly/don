package coms

import "strconv"

import (
	. "don/core"
)

type ProdCom struct{}

func (ProdCom) OutputType(inputType DType) (outputType DType, impossible bool) {
	if inputType.Tag == UnknownTypeTag {
		return
	}
	if inputType.Tag != StructTypeTag {
		impossible = true
		return
	}

	outputType = UnitType
	for i := len(inputType.Fields) - 1; i >= 0; i-- {
		fieldType, fieldExists := inputType.Fields[strconv.FormatInt(int64(i), 10)]
		if !fieldExists {
			impossible = true
			return
		}

		switch fieldType.Tag {
		case UnknownTypeTag:
			outputType = UnknownType
		case UnitTypeTag:
		case StructTypeTag:
			fields := make(map[string]DType, len(fieldType.Fields))
			for fieldName, fieldType := range fieldType.Fields {
				if fieldType.Tag != UnknownTypeTag && fieldType.Tag != UnitTypeTag {
					impossible = true
					return
				}
				fields[fieldName] = outputType
			}
			outputType = MakeStructType(fields)
		}
	}
	return
}

func getFieldName(fieldChan chan<- string, fieldName string, unitChan <-chan Unit, quit <-chan struct{}) {
	for {
		select {
		case <-unitChan:
			fieldChan <- fieldName
		case <-quit:
			return
		}
	}
}

func (ProdCom) Run(inputType DType, inputGetter InputGetter, outputGetter OutputGetter, quit <-chan struct{}) {
	var fieldChans []<-chan string
	var unitChans []<-chan Unit

	for i := 0; i < len(inputType.Fields); i++ {
		fieldName := strconv.FormatInt(int64(i), 10)
		fieldType := inputType.Fields[fieldName]

		input := inputGetter.Struct[fieldName].GetInput(fieldType)
		if inputType.Fields[fieldName].Tag == UnitTypeTag {
			unitChans = append(unitChans, input.Unit)
		} else { // inputType.Fields[fieldName].Tag == StructTypeTag
			fieldChan := make(chan string)
			fieldChans = append(fieldChans, fieldChan)
			for fieldName, fieldInput := range input.Struct {
				go getFieldName(fieldChan, fieldName, fieldInput.Unit, quit)
			}
		}
	}

	outputType, _ := ProdCom{}.OutputType(inputType)
	output := outputGetter.GetOutput(outputType)

	for {
		for _, unitChan := range unitChans {
			select {
			case <-unitChan:
			case <-quit:
				return
			}
		}

		currentOutput := output

		for _, fieldChan := range fieldChans {
			select {
			case fieldName := <-fieldChan:
				currentOutput = currentOutput.Struct[fieldName]
			case <-quit:
				return
			}
		}

		currentOutput.WriteUnit()
	}
}
