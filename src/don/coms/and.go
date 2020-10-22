package coms

import (
	. "don/core"
)

var And Com = PipeCom([]Com{
	ProdCom{},
	SplitMergeCom([]Com{
		PipeCom([]Com{SelectCom("T"), SelectCom("T"), DeselectCom("T")}),
		PipeCom([]Com{
			SplitMergeCom([]Com{
				PipeCom([]Com{SelectCom("T"), SelectCom("F")}),
				PipeCom([]Com{SelectCom("F"), SelectCom("T")}),
				PipeCom([]Com{SelectCom("F"), SelectCom("F")})}),
			DeselectCom("F")})})})
