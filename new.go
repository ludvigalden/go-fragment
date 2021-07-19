package fragment

import (
	"fmt"
	"strings"

	"github.com/ludvigalden/go-typemeta"
)

// NewStruct returns a new fragment
func NewStruct(t interface{}) Struct {
	f, err := ParseStruct(t, nil)
	if err != nil {
		panic(err)
	}
	return f
}

// NewCompleteStruct returns a new complete fragment
func NewCompleteStruct(t interface{}) Struct {
	f := NewEmptyStruct(t)
	f.TypeMeta().IterateFields(func(structField typemeta.StructField) {
		f.fields[structField.Index] = StructField{StructField: structField, Fragment: NewStruct(structField.TypeMeta)}
	})
	return f
}

// NewEmptyStruct returns a new empty struct fragment
func NewEmptyStruct(t interface{}) Struct {
	f, err := ParseStruct(t, nil)
	if err != nil {
		panic(err)
	}
	f.fields = map[int]StructField{}
	return f
}

// NewUnstructured returns a new unstructured fragment. If one or more fields are passed,
// the fragment will explicitly contain those fields.
func NewUnstructured() Unstructured {
	return Unstructured{}
}

// NewEmptyUnstructured returns a new empty unstructured fragment. If one or more fields are passed,
// the fragment will explicitly contain those fields.
func NewEmptyUnstructured() Unstructured {
	return Unstructured{map[string]Fragment{}}
}

// NewStructPath creates a new field path
func NewStructPath(t interface{}, v ...interface{}) StructPath {
	typeMeta := typemeta.StructOf(typemeta.Get(t))
	if typeMeta == nil {
		panic("cannot create struct path for non-struct type " + typemeta.Get(t).String())
	}
	path := StructPath{typeMeta: typeMeta, fieldIndices: []int{}}
	currentTypeMeta := typeMeta
	for _, v := range v {
		if vstr, ok := v.(string); ok {
			v = strings.Split(vstr, ".")
		}
		if vint, ok := v.(int); ok {
			v = []int{vint}
		}
		switch v := v.(type) {
		case []string:
			for _, fieldName := range v {
				field := currentTypeMeta.EnsureFieldByName(fieldName)
				path.fieldIndices = append(path.fieldIndices, field.Index)
				currentTypeMeta = typemeta.StructOf(field.TypeMeta)
			}
			break
		case []int:
			for _, fieldIndex := range v {
				field := currentTypeMeta.EnsureField(fieldIndex)
				path.fieldIndices = append(path.fieldIndices, field.Index)
				currentTypeMeta = typemeta.StructOf(field.TypeMeta)
			}
			break
		case StructPath:
			if v.typeMeta.Type() != currentTypeMeta.Type() {
				panic("expected struct path for type " + currentTypeMeta.String() + " but received " + v.typeMeta.String())
			}
			path.fieldIndices = append(path.fieldIndices, v.fieldIndices...)
			currentTypeMeta = v.TailTypeMeta()
			break
		case Path:
			for _, fieldName := range v.FieldNames() {
				field := currentTypeMeta.EnsureFieldByName(fieldName)
				path.fieldIndices = append(path.fieldIndices, field.Index)
				currentTypeMeta = typemeta.StructOf(field.TypeMeta)
			}
			break
		case Fragment:
			tailFragment, err := ParseStruct(path.TailTypeMeta(), v)
			if err != nil {
				panic(err)
			}
			path.tailFragment = tailFragment
			break
		default:
			panic("unrecognized path argument " + fmt.Sprint(v))
		}
	}
	if len(path.fieldIndices) == 0 {
		path.fieldIndices = nil
	}
	return path
}

// NewUnstructuedPath creates a new field path
func NewUnstructuedPath(v ...interface{}) UnstructuredPath {
	path := UnstructuredPath{fieldNames: []string{}}
	for _, v := range v {
		if vstr, ok := v.(string); ok {
			v = strings.Split(vstr, ".")
		}
		switch v := v.(type) {
		case []string:
			for _, fieldName := range v {
				path.fieldNames = append(path.fieldNames, fieldName)
			}
			break
		case Path:
			path.fieldNames = append(path.fieldNames, v.FieldNames()...)
			break
		case Fragment:
			path.tailFragment = v
		default:
			panic("unrecognized path argument " + fmt.Sprint(v))
		}
	}
	return path
}
