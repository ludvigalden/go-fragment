package fragment

import (
	"errors"

	"github.com/ludvigalden/go-typemeta"
)

// Parse parses a value and turns it into a fragment for a certain type.
func Parse(t interface{}, f interface{}) (Struct, error) {
	typeMeta := typemeta.NonPtr(typemeta.Get(t))
	structTypeMeta := typemeta.StructOf(typeMeta)
	emptyFragment := Struct{typeMeta: structTypeMeta}
	switch f := f.(type) {
	case Struct:
		if typeMeta.Primitive() {
			if !f.IsUndefined() {
				return Struct{}, errors.New("expected undefined fragment for primitive type, but received \"" + f.Expr() + "\"")
			}
			return emptyFragment, nil
		}
		if structTypeMeta == nil || (f.typeMeta != nil && f.typeMeta.Type() != structTypeMeta.Type()) {
			return emptyFragment, errors.New("type mismatch for passed type and type of passed struct fragment: \"" + typeMeta.String() + "\" vs. " + "\"" + f.TypeMeta().String() + "\"")
		}
		if f.typeMeta == nil {
			f.typeMeta = structTypeMeta
		}
		return f, nil
	case Unstructured:
		if typeMeta.Primitive() {
			if !f.IsUndefined() {
				return emptyFragment, errors.New("expected undefined fragment for primitive type, but received \"" + f.Expr() + "\"")
			}
			return emptyFragment, nil
		}
		if structTypeMeta == nil {
			if !f.IsUndefined() {
				return emptyFragment, errors.New("expected undefined fragment for non-fragmentable type, but received \"" + f.Expr() + "\"")
			}
			return emptyFragment, nil
		}
		return ParseStruct(structTypeMeta, f)
	default:
		unstructured, err := ParseUnstructured(f)
		if err != nil {
			return emptyFragment, err
		}
		if structTypeMeta == nil {
			if !unstructured.IsUndefined() {
				return emptyFragment, errors.New("expected undefined fragment for non-fragmentable type, but received \"" + unstructured.Expr() + "\"")
			}
			return emptyFragment, nil
		}
		return unstructured.ToStruct(structTypeMeta)
	}
}
