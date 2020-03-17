package coms

import (
	. "don/core"
)

var And Com = Pipe([]Com{
	ProdCom{},
	SplitMerge([]Com{
		Pipe([]Com{SelectCom("true"), SelectCom("true"), Deselect("true")}),
		Pipe([]Com{
			SplitMerge([]Com{
				Pipe([]Com{SelectCom("true"), SelectCom("false")}),
				Pipe([]Com{SelectCom("false"), SelectCom("true")}),
				Pipe([]Com{SelectCom("false"), SelectCom("false")})}),
			Deselect("false")})})})
