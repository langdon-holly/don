package types

import . "don/core"

func MakeLinkedListType(elementType DType) (ret DType, ret1 DType) {
	ret1 = MakeStructType(make(map[string]DType, 2))

	ret = MakeRefType(ret1)

	ret1.Fields["head"] = elementType
	ret1.Fields["tail"] = ret

	return
}
