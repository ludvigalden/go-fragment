package fragment

import (
	"fmt"
	"reflect"

	"github.com/ludvigalden/go-typemeta"
)

// ParseStruct returns a new fragment for the specified type. The specified type can be a map, slice, or array of structs, or a struct.
// The specified fragment value can be a `fragment.Fragment`, a `fragment.Interface`, or anything that can be parsed by `fragment.ParseUnstructured`.
// That is, a list of strings, a string-interface map, or a string fragment string such as "fullName, profile { createdAt }".
// If the specified type meta is not for a struct (or if its element is not a struct), a undefined fragment is returned.
func ParseStruct(t interface{}, f ...interface{}) (Struct, error) {
	typeMeta := typemeta.Get(t)

	// check if the type is fragmentable. if not, we create an error to return if any fragment argument is passed that is non-nil
	structTypeMeta := typemeta.StructOf(typeMeta)
	var noStructErr error
	structf := Struct{typeMeta: structTypeMeta}
	if structTypeMeta == nil {
		noStructErr = NewError("cannot fragment unfragmentable type " + fmt.Sprint(typemeta.StructOrPrimitiveOf(typeMeta)) + " (" + typeMeta.String() + ", " + fmt.Sprint(f...) + ")")
	}
	flen := len(f)
	if flen == 0 {
		return structf, nil
	}
	for _, f := range f {
		if f == nil {
			continue
		}
		if vf, ok := f.(Struct); ok {
			if vf.IsUndefined() {
				continue
			} else if noStructErr != nil {
				return vf, noStructErr
			}
			if vf.TypeMeta().Type() != structTypeMeta.Type() {
				return vf, NewError("type mismatch between passed struct fragment and struct type meta: " + vf.TypeMeta().String() + " vs. " + structTypeMeta.String())
			}
			if flen == 1 {
				return vf, nil
			}
			structf = structf.assign(vf)
			continue
		}
		fi, err := ParseUnstructured(f)
		if err != nil {
			return structf, err
		} else if fi.IsUndefined() {
			continue
		} else if noStructErr != nil {
			return structf, noStructErr
		}
		fis, err := fi.toStruct(structTypeMeta, map[reflect.Type]bool{})
		if err != nil {
			return structf, err
		}
		if flen == 1 {
			structf = fis
		} else {
			structf = structf.assign(fis)
		}
	}
	return structf, nil
}

// IsFragmentable returns whether a struct fragment can be created for the specified type.
// It returns true if `typemeta.StructOf` does not return nil and the struct is not a primitive struct.
func IsFragmentable(t interface{}) bool {
	typeMeta := typemeta.Get(t)
	if typeMeta == nil {
		return false
	}
	structTypeMeta := typemeta.StructOf(typeMeta)
	if structTypeMeta == nil {
		return false
	} else if structTypeMeta.Primitive() {
		return false
	}
	return true
}
