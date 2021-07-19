package fragment

import (
	"errors"
	"reflect"

	"github.com/ludvigalden/go-typemeta"
)

// SetValueTo sets
func (sp StructPath) SetValueTo(value reflect.Value, out reflect.Value) (reflect.Value, error) {
	nonPtrOut := out
	for nonPtrOut.Kind() == reflect.Ptr {
		if nonPtrOut.IsNil() {
			return out, errors.New("unable to set field path value to nil pointer")
		}
		nonPtrOut = nonPtrOut.Elem()
	}
	switch nonPtrOut.Kind() {
	case reflect.Slice:
		elemType := nonPtrOut.Type().Elem()
		appended := nonPtrOut
		var err error
		sp.IterateValues(value, func(pathValue reflect.Value) {
			if IsValueJSONNull(pathValue) {
				return
			}
			parsedElem, convErr := typemeta.ConvertValue(pathValue, elemType)
			if convErr != nil {
				err = convErr
				return
			}
			appended = reflect.Append(nonPtrOut, parsedElem)
		})
		if err != nil {
			return out, err
		}
		nonPtrOut.Set(appended)
	default:
		found := sp.FindValue(value, func(pathValue reflect.Value) bool {
			return !IsValueJSONNull(pathValue)
		})
		if found == nil {
			return out, nil
		}
		parsedPathValue, err := typemeta.ConvertValue(*found, nonPtrOut.Type())
		if err != nil {
			return out, err
		}
		nonPtrOut.Set(parsedPathValue)
		break
	}

	return out, nil
}

// FindValue iterates all values at the path of the field. The type of the value must equal the type specified for the path.
// The iteratee can return `true` to stop the iteration (implying that the value has been found), and in that case the value iterated at that point will be returned.
func (sp StructPath) FindValue(value reflect.Value, iteratee func(pathValue reflect.Value) bool) *reflect.Value {
	var found *reflect.Value
	iterateValuesAtPath(value, sp.typeMeta, sp.FieldIndices(), func(structField typemeta.StructField, pathValue reflect.Value) bool {
		if iteratee(pathValue) {
			found = &pathValue
			return true
		}
		return false
	})
	return found
}

// FindFieldValue iterates all values at the path of the field. The type of the value must equal the type specified for the path.
// The iteratee can return `true` to stop the iteration (implying that the value has been found), and in that case the value iterated at that point will be returned.
func (sp StructPath) FindFieldValue(value reflect.Value, iteratee func(structField typemeta.StructField, pathValue reflect.Value) bool) *reflect.Value {
	var found *reflect.Value
	iterateValuesAtPath(value, sp.typeMeta, sp.FieldIndices(), func(structField typemeta.StructField, pathValue reflect.Value) bool {
		if iteratee(structField, pathValue) {
			found = &pathValue
			return true
		}
		return false
	})
	return found
}

// IterateValues iterates all values at the path of the field. The type of the value must equal the type specified for the path.
func (sp StructPath) IterateValues(value reflect.Value, iteratee func(pathValue reflect.Value)) {
	iterateValuesAtPath(value, sp.typeMeta, sp.FieldIndices(), func(structField typemeta.StructField, pathValue reflect.Value) bool {
		iteratee(pathValue)
		return false
	})
}

// IterateFieldValues iterates all values at the path of the field. The type of the value must equal the type specified for the path.
func (sp StructPath) IterateFieldValues(value reflect.Value, iteratee func(structField typemeta.StructField, pathValue reflect.Value)) {
	iterateValuesAtPath(value, sp.typeMeta, sp.FieldIndices(), func(structField typemeta.StructField, pathValue reflect.Value) bool {
		iteratee(structField, pathValue)
		return false
	})
}

func iterateValuesAtPath(value reflect.Value, typeMeta typemeta.TypeMeta, path []int, iteratee func(structField typemeta.StructField, pathValue reflect.Value) bool) bool {
	pathLen := len(path)
	if pathLen == 0 {
		iteratee(typemeta.StructField{}, value)
		return false
	}
	nonPtrValue := value
	for nonPtrValue.Kind() == reflect.Ptr {
		if nonPtrValue.IsNil() {
			return false
		}
		nonPtrValue = nonPtrValue.Elem()
	}
	typeMeta = typemeta.NonPtr(typeMeta)
	if nonPtrValue.Type() != typeMeta.Type() {
		panic("expected type \"" + typeMeta.Type().String() + "\", but received \"" + nonPtrValue.Type().String() + "\"")
	}

	switch nonPtrValue.Kind() {
	case reflect.Struct:
		structTypeMeta := typemeta.StructOf(typeMeta)
		if structTypeMeta == nil {
			panic("received struct value for non-struct type")
		}
		nextFieldIndex := path[0]
		nextFieldValue := nonPtrValue.Field(nextFieldIndex)
		if pathLen > 1 {
			if iterateValuesAtPath(nextFieldValue, structTypeMeta.Field(nextFieldIndex).TypeMeta, path[1:], iteratee) {
				return true
			}
		} else {
			if iteratee(structTypeMeta.EnsureField(nextFieldIndex), nextFieldValue) {
				return true
			}
		}
		break
	case reflect.Slice:
		sliceTypeMeta := typemeta.AssertSlice(typeMeta)
		if sliceTypeMeta == nil {
			panic("received slice value for non-slice type")
		}
		for index := 0; index < nonPtrValue.Len(); index++ {
			if iterateValuesAtPath(nonPtrValue.Index(index), sliceTypeMeta.Elem, path, iteratee) {
				return true
			}
		}
		break
	case reflect.Array:
		arrayTypeMeta := typemeta.AssertArray(typeMeta)
		if arrayTypeMeta == nil {
			panic("received array value for non-array type")
		}
		for index := 0; index < nonPtrValue.Len(); index++ {
			if iterateValuesAtPath(nonPtrValue.Index(index), arrayTypeMeta.Elem, path, iteratee) {
				return true
			}
		}
		break
	case reflect.Map:
		mapTypeMeta := typemeta.MapOf(typeMeta)
		if mapTypeMeta == nil {
			panic("received map value for non-map type")
		}
		mapIter := nonPtrValue.MapRange()
		for mapIter.Next() {
			if iterateValuesAtPath(mapIter.Value(), mapTypeMeta.Elem, path, iteratee) {
				return true
			}
		}
		break
	default:
		panic("expected type \"" + typeMeta.String() + "\", cannot iterate path")
	}
	return false
}
