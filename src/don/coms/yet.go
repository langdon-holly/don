package coms

import (
	. "don/core"
	"don/types"
)

var yetComInputType = MakeNFieldsType(2)

func init() {
	yetComInputType.Fields[""] = UnitType
	yetComInputType.Fields["?"] = UnitType
}

func Yet() Com { return YetCom{} }

type YetCom struct{}

func (YetCom) InputType() DType  { return yetComInputType }
func (YetCom) OutputType() DType { return types.BoolType }

func (YetCom) MeetTypes(inputType, outputType DType) Com {
	if !yetComInputType.LTE(inputType) || !types.BoolType.LTE(outputType) {
		return Null
	} else {
		return YetCom{}
	}
}

func (YetCom) Underdefined() Error { return nil }

func (YetCom) Copy() Com { return YetCom{} }

func (YetCom) Invert() Com { return InverseYetCom{} }

func (YetCom) Run(input Input, output Output) {
	<-input.Fields["?"].Unit
	select {
	case <-input.Fields[""].Unit:
		output.Fields["T"].Converge()
	default:
		output.Fields["F"].Converge()
	}
}
