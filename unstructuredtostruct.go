package fragment

import (
	"errors"
	"reflect"
	"sync"

	"github.com/ludvigalden/go-typemeta"
)

// ToStruct validates the fragment and returns a result
func (f Unstructured) ToStruct(typeMeta *typemeta.Struct) (Struct, error) {
	return f.toStruct(typeMeta, map[reflect.Type]bool{})
}

func (f Unstructured) toStruct(typeMeta *typemeta.Struct, circular map[reflect.Type]bool) (Struct, error) {
	result := Struct{typeMeta: typeMeta}
	if typeMeta == nil {
		return result, errors.New("received nil type meta")
	} else if f.fields == nil {
		return result, nil
	}
	if f.fields != nil && len(f.fields) > 0 {
		if typeMeta == nil {
			panic("expected fragmentable type meta for non-undefined fragment")
		}
		result.fields = map[int]StructField{}
		unrecognizedFields := []string{}
		for fieldName, fieldFragment := range f.fields {
			structField := typeMeta.FieldByName(fieldName)
			if structField == nil {
				if fieldName != "__typename" {
					unrecognizedFields = append(unrecognizedFields, fieldName)
				}
				continue
			}
			if fieldFragment == nil || fieldFragment.IsUndefined() {
				result.fields[structField.Index] = StructField{StructField: *structField}
			} else {
				structFieldStructTypeMeta := typemeta.StructOf(structField.TypeMeta)
				if structFieldStructTypeMeta == nil {
					return result, errors.New("expected undefined fragment for non-fragmentable field \"" + structField.String() + "\"")
				}
				fieldFragment, err := ParseStruct(structFieldStructTypeMeta, fieldFragment)
				if err != nil {
					return result, errors.New("invalid fragment for field \"" + structField.String() + "\": " + err.Error())
				}
				result.fields[structField.Index] = StructField{StructField: *structField, Fragment: fieldFragment}
			}
		}
		if len(unrecognizedFields) > 0 {
			return result, errors.New("unrecognized field(s): " + fmtListAnd("en", quotedStringInteraces(unrecognizedFields...)...))
		}
	} else if typeMeta != nil {
		if circular[typeMeta.Type()] {
			return result, nil
		}
		result.fields = map[int]StructField{}
		circular[typeMeta.Type()] = true
		var fieldErr error
		hasIncludeDefaults := typeHasIncludeDefaults(typeMeta)
		typeMeta.IterateFields(func(structField typemeta.StructField) {
			if hasIncludeDefaults && !includeDefault(structField) {
				return
			}
			fieldStructTypeMeta := typemeta.StructOf(structField.TypeMeta)
			if fieldStructTypeMeta != nil {
				fieldFragment, err := NewUnstructured().toStruct(fieldStructTypeMeta, circular)
				if err != nil {
					fieldErr = errors.New("invalid fragment for field \"" + structField.String() + "\": " + err.Error())
				}
				result.fields[structField.Index] = StructField{StructField: structField, Fragment: fieldFragment}
			} else {
				result.fields[structField.Index] = StructField{StructField: structField}
			}
		})
		if fieldErr != nil {
			return result, fieldErr
		}
	}
	return result, nil
}

func includeDefault(structField typemeta.StructField) bool {
	fragmentTag := structField.Tag("fragment")
	return fragmentTag != nil && fragmentTag.Name == "includedefault"
}

func typeHasIncludeDefaults(typeMeta typemeta.TypeMeta) bool {
	structTypeMeta := typemeta.StructOf(typeMeta)
	if structTypeMeta == nil {
		return false
	}
	nonPtrType := typeMeta.Type()
	hasIncludeDefaultMutex.Lock()
	hid, ok := hasIncludeDefaultCache[nonPtrType]
	if ok {
		hasIncludeDefaultMutex.Unlock()
		return hid
	}
	hid = structTypeMeta.FindField(func(structField typemeta.StructField) bool {
		return includeDefault(structField)
	}) != nil
	hasIncludeDefaultCache[nonPtrType] = hid
	hasIncludeDefaultMutex.Unlock()
	return hid
}

var hasIncludeDefaultCache = map[reflect.Type]bool{}
var hasIncludeDefaultMutex sync.Mutex
