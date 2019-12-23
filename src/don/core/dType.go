package core

// DType

type DTypeTag int

const (
	UnitTypeTag = DTypeTag(iota)
	RefTypeTag
	StructTypeTag
)

type DTypeLvl int

const (
	UnknownTypeLvl = DTypeLvl(iota)
	NormalTypeLvl
	ImpossibleTypeLvl
)

type DType struct {
	Lvl      DTypeLvl
	Tag      DTypeTag         /* for Lvl == NormalTypeLvl */
	Referent *DType           /* for Lvl == NormalTypeLvl && Tag == RefTypeTag */
	Fields   map[string]DType /* for Lvl == NormalTypeLvl && Tag == StructTypeTag */
}

// What TODO about partial struct types?

// Get DType

var UnknownType = DType{}

var ImpossibleType = DType{Lvl: ImpossibleTypeLvl}

var UnitType = DType{Lvl: NormalTypeLvl, Tag: UnitTypeTag}

func MakeRefType(referentType DType) DType {
	return DType{Lvl: NormalTypeLvl, Tag: RefTypeTag, Referent: &referentType}
}

func MakeStructType(fields map[string]DType) DType {
	return DType{Lvl: NormalTypeLvl, Tag: StructTypeTag, Fields: fields}
}

func TypeAtPath(theType DType, fieldPath []string) DType {
	for i := len(fieldPath) - 1; i >= 0; i-- {
		fields := make(map[string]DType, 1)
		fields[fieldPath[i]] = theType

		theType = MakeStructType(fields)
	}
	return theType
}

// Other

func assumingEqual(t0, t1 DType, assumedEquals map[*DType]map[*DType]struct{}) bool {
	if t0.Lvl != t1.Lvl {
		return false
	}
	if t0.Lvl != NormalTypeLvl {
		return true
	}

	if t0.Tag != t1.Tag {
		return false
	}

	if t0.Tag == RefTypeTag {
		rights, ok := assumedEquals[t0.Referent]
		if !ok {
			rights = make(map[*DType]struct{}, 1)
			assumedEquals[t0.Referent] = rights
		}

		_, ok = rights[t1.Referent]
		if ok {
			/* assumed equal */
			return true
		} else {
			/* assume they're equal */
			rights[t1.Referent] = struct{}{}
			return assumingEqual(*t0.Referent, *t1.Referent, assumedEquals)
		}
	} else if t0.Tag == StructTypeTag {

		if len(t0.Fields) != len(t1.Fields) {
			return false
		}
		for fieldName, fieldType0 := range t0.Fields {
			fieldType1, exists := t1.Fields[fieldName]
			if !exists {
				return false
			}
			if !fieldType0.Equal(fieldType1) {
				return false
			}
		}
	}

	return true
}

func (t0 DType) Equal(t1 DType) bool {
	return assumingEqual(t0, t1, make(map[*DType]map[*DType]struct{}))
}

func recursiveMerge(t0, t1 DType, referentMerges map[*DType]map[*DType]*DType) DType {
	if t0.Lvl == UnknownTypeLvl {
		return t1
	}
	if t1.Lvl == UnknownTypeLvl {
		return t0
	}
	if t0.Lvl == ImpossibleTypeLvl || t1.Lvl == ImpossibleTypeLvl {
		return ImpossibleType
	}

	// t0.Lvl == t1.Lvl == NormalTypeLvl

	if t0.Tag != t1.Tag {
		return ImpossibleType
	}
	switch t0.Tag {
	case UnitTypeTag:
		return t0
	case RefTypeTag:
		merged := DType{Lvl: NormalTypeLvl, Tag: RefTypeTag}

		rights, ok := referentMerges[t0.Referent]
		if !ok {
			rights = make(map[*DType]*DType, 1)
			referentMerges[t0.Referent] = rights
		}

		if merge, inProgress := rights[t1.Referent]; inProgress {
			merged.Referent = merge.Referent
		} else {
			merged.Referent = new(DType)
			rights[t1.Referent] = merged.Referent
			*merged.Referent = recursiveMerge(*t0.Referent, *t1.Referent, referentMerges)
		}

		return merged
	case StructTypeTag:
		if len(t0.Fields) != len(t1.Fields) {
			return ImpossibleType
		}

		ret := DType{Lvl: NormalTypeLvl, Tag: StructTypeTag, Fields: make(map[string]DType, len(t0.Fields))}

		for fieldName, t0FieldType := range t0.Fields {
			t1FieldType, exists := t1.Fields[fieldName]
			if !exists {
				return ImpossibleType
			}
			ret.Fields[fieldName] = MergeTypes(t0FieldType, t1FieldType)
		}
		return ret
	default:
		panic("Unreachable")
	}
}

func MergeTypes(t0, t1 DType) DType {
	return recursiveMerge(t0, t1, make(map[*DType]map[*DType]*DType))
}
