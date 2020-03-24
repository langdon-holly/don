package coms

import "strconv"

import . "don/core"

// len(pipes) >= 1
func ComCom(pipes []Com) Com {
	if len(pipes) == 1 {
		return pipes[0]
	}

	// len(pipes) >= 2

	indexStrings := make([]string, len(pipes)-1)
	comEntries := make([]CompositeComEntry, len(pipes)+1)

	comEntries[len(pipes)].OutputMap = SignalMap{
		ParentP:  true,
		Children: make(map[string]SignalMap, len(pipes)-1)}

	comEntries[0].OutputMap.ExternalP = true
	{
		i := 0
		indexString := "0"
		goto first
	next:
		indexString = strconv.FormatInt(int64(i), 10)

		comEntries[i].OutputMap = SignalMap{
			InternalIdx: len(pipes),
			FieldPath:   []string{indexString}}
	first:
		comEntries[i].Com = pipes[i]

		if i < len(pipes)-1 {
			indexStrings[i] = indexString
			comEntries[len(pipes)].OutputMap.Children[indexString] =
				SignalMap{InternalIdx: i}

			i++
			goto next
		}
	}

	comEntries[len(pipes)].Com = Pipe([]Com{MergeCom{}, SplitCom(indexStrings)})

	inputMap := SignalMap{InternalIdx: len(pipes) - 1}

	return CompositeCom{Coms: comEntries, InputMap: inputMap}
}
