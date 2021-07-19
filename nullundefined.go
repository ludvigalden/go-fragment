package fragment

import (
	"reflect"

	"github.com/ludvigalden/go-typemeta"
)

// IsValueUndefined returns whether a value is deemed to be undefined in the opinionated view of this package.
// In this view, number is zero when it is possible that it was not specified by the user or specified by the user as undefined.
// In this view, number and boolean values are special, because there is a meaningful difference between `nil`, `0`, and `false`.
// Therefore, a pointer to `0` or `false` is deemed defined, while a non-pointer value `0` or `false` is deemed undefined.
// Otherwise, (reflect.Value).IsZero() is used.
func IsValueUndefined(v reflect.Value) bool {
	return isValueUndefined(v, false)
}

func isValueUndefined(v reflect.Value, elem bool) bool {
	switch v.Kind() {
	case reflect.String, reflect.Struct:
		if elem {
			return false
		}
		return v.IsZero()
	case reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return isValueUndefined(v.Elem(), true)
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if elem {
			return false
		}
		return v.IsZero()
	case reflect.Invalid:
		return true
	default:
		return v.IsZero()
	}
}

// IsValueJSONNull returns whether a value is deemed to be presented as null in JSON-format in the opinionated view of this package.
// The following values are considered JSON-null: (1) nil pointers, (2) pointers to JSON-null values, (3) empty strings, (4) slices, arrays, or maps that are empty or only containing JSON-null elements,
// and (5) invalid values.
func IsValueJSONNull(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return IsValueJSONNull(v.Elem())
	case reflect.Invalid:
		return true
	case reflect.Slice, reflect.Array:
		if v.IsNil() {
			return true
		}
		for i := 0; i < v.Len(); i++ {
			if !IsValueJSONNull(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Map:
		if v.IsNil() {
			return true
		}
		for _, key := range v.MapKeys() {
			if !IsValueJSONNull(v.MapIndex(key)) {
				return false
			}
		}
		return true
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return false
	default:
		return v.IsZero()
	}
}

// IsFieldValueJSONNull returns whether a value is deemed to be presented as null in JSON-format in the opinionated view of this package.
// The following values are considered JSON-null: (1) nil pointers, (2) pointers to JSON-null values, (3) empty strings, (4) slices, arrays, or maps that are empty or only containing JSON-null elements,
// and (5) invalid values.
func IsFieldValueJSONNull(structField *typemeta.StructField, v reflect.Value) bool {
	if structField == nil {
		return IsValueJSONNull(v)
	}
	return IsValueJSONNull(v) || (structField.JSONOmitEmpty && IsValueUndefined(v))
}
