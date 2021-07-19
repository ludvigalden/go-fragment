package fragment

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/ludvigalden/go-typemeta"
)

// PickJSON returns value to be marshaled using `json.Marshal`, e.g. a `map[string]interface{}`.
func PickJSON(fragment Fragment, value interface{}) (interface{}, error) {
	if fragment == nil {
		return value, nil
	}
	reflectValue, err := pickJSON(fragment, reflect.ValueOf(value))
	if !reflectValue.IsValid() {
		return nil, err
	}
	return reflectValue.Interface(), err
}

// MarshalJSON uses `PickJSON` to get a value to marshal and then marshals it using `json.Marshal`.
func MarshalJSON(fragment Fragment, value interface{}) ([]byte, error) {
	picked, err := PickJSON(fragment, value)
	if err != nil {
		return nil, err
	}
	return json.Marshal(picked)
}

func pickJSON(fragment Fragment, reflectValue reflect.Value) (reflect.Value, error) {
	var newReflectValue reflect.Value
	if fragment == nil {
		return reflectValue, nil
	}
	if reflectValue.Kind() == reflect.Interface {
		reflectValue = reflect.ValueOf(reflectValue.Interface())
	}
	nonPtrReflectValue := reflectValue
	if reflectValue.Kind() == reflect.Ptr {
		if reflectValue.IsNil() {
			return newReflectValue, nil
		}
		nonPtrReflectValue = reflectValue.Elem()
	}
	if nonPtrReflectValue.Kind() == reflect.Array || nonPtrReflectValue.Kind() == reflect.Slice {
		jsonSlice := []interface{}{}
		for index := 0; index < nonPtrReflectValue.Len(); index++ {
			json, err := pickJSON(fragment, nonPtrReflectValue.Index(index))
			if err != nil {
				return newReflectValue, err
			}
			if !json.IsValid() {
				jsonSlice = append(jsonSlice, nil)
			} else {
				jsonSlice = append(jsonSlice, json.Interface())
			}
		}
		newReflectValue = reflect.ValueOf(jsonSlice)
		return newReflectValue, nil
	} else if fragment, ok := fragment.(Struct); ok {
		// fragment = fragment.EnsureDefined(reflectValue.Type())
		if nonPtrReflectValue.Kind() == reflect.Struct {
			if fragment.TypeMeta().Primitive() {
				return nonPtrReflectValue, nil
			}
			structTypeMeta := typemeta.StructOf(fragment.TypeMeta())
			if nonPtrReflectValue.Type() != structTypeMeta.Type() {
				return reflectValue, errors.New("type of value and fragment do not match: " + nonPtrReflectValue.Type().String() + " vs. " + fragment.TypeMeta().String())
			}
			values := map[string]interface{}{}
			for index := 0; index < nonPtrReflectValue.NumField(); index++ {
				if !fragment.HasByIndex(index) {
					continue
				}
				structField := structTypeMeta.EnsureField(index)
				if structField.JSONName == "" {
					continue
				}
				field := fragment.Field(index)
				fieldOriginalValue := nonPtrReflectValue.Field(index)
				if IsFieldValueJSONNull(&structField, fieldOriginalValue) {
					if !fragment.IsUndefined() {
						// the field was included specifically, so set the value to nil
						values[field.JSONName] = nil
					}
					continue
				}
				if !field.Primitive() && field.Fragment.IsUndefined() {
					field = StructField{StructField: structField, Fragment: NewStruct(structField.TypeMeta)}
				}
				fieldValue := fieldOriginalValue
				if typemeta.StructOf(field.TypeMeta) != nil || typemeta.InterfaceOf(field.TypeMeta) != nil {
					var err error
					fieldValue, err = pickJSON(field.Fragment, fieldOriginalValue)
					if err != nil {
						return reflect.Value{}, NewError(err).Register(structField.Name)
					}
					if IsValueJSONNull(fieldValue) {
						if !fragment.IsUndefined() {
							// the field was included specifically, so set the value to nil
							values[field.JSONName] = nil
						}
						continue
					}
				}
				values[field.JSONName] = fieldValue.Interface()
			}
			newReflectValue = reflect.ValueOf(values)
		} else {
			newReflectValue = nonPtrReflectValue
		}
		return newReflectValue, nil
	} else if nonPtrReflectValue.Kind() == reflect.Struct {
		return nonPtrReflectValue, errors.New("not implemented picking structs using unstructured fragments")
	} else {
		newReflectValue = nonPtrReflectValue
	}
	return newReflectValue, nil
}
