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

func Yet() Com {
	return &YetCom{inputType: yetComInputType, outputType: types.BoolType}
}

type YetCom struct{ inputType, outputType DType }

func (yc *YetCom) InputType() *DType  { return &yc.inputType }
func (yc *YetCom) OutputType() *DType { return &yc.outputType }

func (yc *YetCom) Types() Com {
	if !yetComInputType.LTE(yc.inputType) || !types.BoolType.LTE(yc.outputType) {
		return Null
	} else {
		return yc
	}
}

func (yc YetCom) Underdefined() Error { return nil }

func (yc YetCom) Copy() Com { return &yc }

func (yc *YetCom) Invert() Com { return InverseYetCom{Yet: yc} }

func (yc YetCom) Run(input Input, output Output) {
	if !yetComInputType.LTE(yc.inputType) || !types.BoolType.LTE(yc.outputType) {
		return
	}
	<-input.Fields["?"].Unit
	select {
	case <-input.Fields[""].Unit:
		output.Fields["T"].Converge()
	default:
		output.Fields["F"].Converge()
	}
}
