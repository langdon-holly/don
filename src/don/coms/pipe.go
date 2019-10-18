package coms

import . "don/core"

// len(coms) > 0
func pipe(coms []Com) (ret CompositeCom) {
	ret.TheInputType = coms[0].InputType()
	ret.TheOutputType = coms[len(coms)-1].OutputType()

	ret.ComEntries = make([]CompositeComEntry, len(coms))
	for i, com := range coms {
		ret.ComEntries[i].Com = com
	}

	for i := 0; i < len(ret.ComEntries)-1; i++ {
		//TODO: Check types
		MakeCompositeComMaps(&ret.ComEntries[i].OutputMap, &ret.ComEntries[i+1].InputMap, &ret.InputChanN, ret.ComEntries[i].OutputType())
	}

	ret.OutputChanN = ret.InputChanN
	ret.InnerChanN = ret.InputChanN

	MakeCompositeComMaps(&ret.InputMap, &ret.ComEntries[0].InputMap, &ret.InputChanN, ret.TheInputType)
	MakeCompositeComMaps(&ret.OutputMap, &ret.ComEntries[len(coms)-1].OutputMap, &ret.OutputChanN, ret.TheOutputType)

	return
}

func GenPipe(genComs []GenCom) GenCom {
	genComEntries := make([]GenCompositeEntry, len(genComs))
	for i := 0; i < len(genComs)-1; i++ {
		genComEntries[i] =
			GenCompositeEntry{
				GenCom: genComs[i],
				OutputMap: SignalReaderIdTree{
					SignalReaderId: SignalReaderId{
						ReaderId: ReaderId{InternalP: true, InternalIdx: i + 1}}}}
	}
	genComEntries[len(genComs)-1].GenCom = genComs[len(genComs)-1]

	inputMap := SignalReaderIdTree{
		SignalReaderId: SignalReaderId{
			ReaderId: ReaderId{InternalP: true}}}

	return GenComposite{GenComs: genComEntries, InputMap: inputMap}
}
