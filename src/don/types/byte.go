package types

import . "don/core"

var ByteType = MakeStructType(make(map[string]DType, 8))

func init() {
	ByteType.Fields["0"] = BoolType
	ByteType.Fields["1"] = BoolType
	ByteType.Fields["2"] = BoolType
	ByteType.Fields["3"] = BoolType
	ByteType.Fields["4"] = BoolType
	ByteType.Fields["5"] = BoolType
	ByteType.Fields["6"] = BoolType
	ByteType.Fields["7"] = BoolType
}
