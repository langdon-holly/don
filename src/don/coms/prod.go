package coms

import "strconv"

import (
	. "don/core"
)

type ProdCom struct{}

func (ProdCom) OutputType(inputType DType) DType {
	switch inputType.Lvl {
	case UnknownTypeLvl:
		return UnknownType
	case NormalTypeLvl:
		if inputType.Tag != StructTypeTag {
			return ImpossibleType
		}

		outputType := UnitType
		for i := len(inputType.Fields) - 1; i >= 0; i-- {
			fieldType, fieldExists := inputType.Fields[strconv.FormatInt(int64(i), 10)]
			if !fieldExists {
				return ImpossibleType
			}

			switch fieldType.Lvl {
			case UnknownTypeLvl:
				outputType = UnknownType
			case NormalTypeLvl:
				switch fieldType.Tag {
				case UnitTypeTag:
				case RefTypeTag:
					return ImpossibleType
				case StructTypeTag:
					fields := make(map[string]DType, len(fieldType.Fields))
					for fieldName, fieldType := range fieldType.Fields {
						if fieldType.Lvl == ImpossibleTypeLvl ||
							fieldType.Lvl == NormalTypeLvl && fieldType.Tag != UnitTypeTag {
							return ImpossibleType
						}
						fields[fieldName] = outputType
					}
					outputType = MakeStructType(fields)
				}
			case ImpossibleTypeLvl:
				return ImpossibleType
			}
		}
		return outputType

	case ImpossibleTypeLvl:
		return ImpossibleType
	}

	panic("Unreachable")
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

	output := outputGetter.GetOutput(ProdCom{}.OutputType(inputType))

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
