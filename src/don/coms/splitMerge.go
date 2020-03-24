package coms

import "strconv"

import . "don/core"

func SplitMerge(coms []Com) Com {
	indexStrings := make([]string, len(coms))
	comEntries := make([]CompositeComEntry, len(coms)+2)

	comEntries[len(coms)].OutputMap = SignalMap{
		ParentP:  true,
		Children: make(map[string]SignalMap, len(coms))}

	for i := 0; i < len(coms); i++ {
		indexString := strconv.FormatInt(int64(i), 10)
		indexStrings[i] = indexString

		comEntries[i] = CompositeComEntry{
			Com: coms[i],
			OutputMap: SignalMap{
				InternalIdx: len(coms) + 1,
				FieldPath:   []string{indexString}}}
		comEntries[len(coms)].OutputMap.Children[indexString] =
			SignalMap{InternalIdx: i}
	}
	comEntries[len(coms)].Com = SplitCom(indexStrings)
	comEntries[len(coms)+1] = CompositeComEntry{
		Com:       MergeCom{},
		OutputMap: SignalMap{ExternalP: true}}

	inputMap := SignalMap{InternalIdx: len(coms)}

	return CompositeCom{Coms: comEntries, InputMap: inputMap}
}
