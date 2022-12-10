package com

//func Swap() Com {
//	t := PairTypePtr()
//	UnifyTypePtrs(GetLeft(t), swapType(GetRight(t)))
//	return SwapCom{T: t}
//}
//
//type SwapCom struct{ T *TypePtr }
//
//func (sc SwapCom) Type() *TypePtr { return sc.T }
//
//func swapType(t *TypePtr) *TypePtr {
//	swapped := AnyTypePtr()
//	if jt, tJunct := TypePtrType(t).(JunctiveType); tJunct {
//		for fieldName, fieldTypePtr := range jt.Juncts {
//			jFieldType, fieldTypeJunct := TypePtrType(fieldTypePtr).(JunctiveType)
//			if !fieldTypeJunct {
//				UnifyTypePtrs(swapped, fieldTypePtr)
//			} else if jt.Junctive == jFieldType.Junctive {
//				for subFieldName, subFieldTypePtr := range jFieldType.Juncts {
//					UnifyTypePtrs(
//						swapped,
//						TypePtrAt(
//							jFieldType.Junctive,
//							subFieldName,
//							TypePtrAt(jt.Junctive, fieldName, subFieldTypePtr),
//						),
//					)
//				}
//			} else if UnifyTypePtrs(swapped, NoTypePtr()); true {
//			}
//		}
//	} else if UnifyTypePtrs(swapped, t); true {
//	}
//	return swapped
//}
//
//func (sc SwapCom) Copy(mapping map[*TypePtr]*TypePtr) Com {
//	return SwapCom{CopyTypePtr(sc.T, mapping)}
//}
//func (sc SwapCom) Convert() Com   { return SwapCom{ConvertTypePtr(sc.T)} }
//func (sc SwapCom) Syntax() Syntax { return NameSyntax("swap") }
//func (sc SwapCom) String() string { return sc.Syntax().String() }
