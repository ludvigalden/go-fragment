package fragment

import (
	"sort"

	"github.com/ludvigalden/go-typemeta"
)

// Struct is a fragment for a specific struct type.
type Struct struct {
	typeMeta *typemeta.Struct `json:"-"`
	fields   map[int]StructField
}

var _ Fragment = Struct{}

// StructField is a fragment field. If the field type is primitive, the field fragment is guaranteed to be undefined.
type StructField struct {
	typemeta.StructField
	// The fragment for the field, which may be undefined (implying that all fields are queried, if the type is non-primitive).
	// Even if the fragment is undefined, methods such as `Has` still works, which returns true.
	Fragment Struct
}

func newStructField(structField typemeta.StructField) StructField {
	return StructField{StructField: structField, Fragment: Struct{typeMeta: typemeta.StructOf(structField.TypeMeta)}}
}

// Add adds the fields with the specified indices to the fragment
func (f Struct) Add(fieldIndices ...int) Struct {
	if len(fieldIndices) == 0 {
		return f
	}
	if f.typeMeta == nil {
		panic("Cannot add to invalid fragment")
	}
	f = f.definedOrEmptyCopy()
	for _, fieldIndex := range fieldIndices {
		if _, ok := f.fields[fieldIndex]; !ok {
			f.fields[fieldIndex] = newStructField(f.typeMeta.EnsureField(fieldIndex))
		}
	}
	return f
}

// AddByName adds the fields with the specified names to the fragment
func (f Struct) AddByName(fieldNames ...string) Struct {
	if len(fieldNames) == 0 || f.typeMeta == nil {
		return f
	}
	f = f.definedOrEmptyCopy()
	for _, fieldName := range fieldNames {
		fieldIndex := f.typeMeta.EnsureFieldByName(fieldName).Index
		if _, ok := f.fields[fieldIndex]; !ok {
			f.fields[fieldIndex] = newStructField(f.typeMeta.EnsureField(fieldIndex))
		}
	}
	return f
}

// AddExcluded adds the specified field to the fragment unless they are already included, as determined by `Has`.
// This differs from `Add` in that `Add` may turn a undefined fragment into a non-undefined fragment, or in other words
// omit fields that was included by default when the fragment was nil.
// `AddExcluded` only adds field that are not (explicitly or implicitly) included.
func (f Struct) AddExcluded(fieldIndices ...int) Struct {
	if len(fieldIndices) == 0 {
		return f
	}
	if f.typeMeta == nil {
		panic("Cannot add to invalid fragment")
	}
	for _, fieldIndex := range fieldIndices {
		structField := f.typeMeta.EnsureField(fieldIndex)
		if f.HasByIndex(structField.Index) {
			continue
		}
		f = f.definedOrEmptyCopy()
		f.fields[structField.Index] = newStructField(structField)
	}
	return f
}

// AddExcludedByName adds the specified field to the fragment unless they are already included, as determined by `Has`.
// This differs from `AddByName` in that `AddByName` may turn a undefined fragment into a non-undefined fragment, or in other words
// omit fields that was included by default when the fragment was nil.
// `AddExcludedByName` only adds field that are not (explicitly or implicitly) included.
func (f Struct) AddExcludedByName(fieldNames ...string) Struct {
	if len(fieldNames) == 0 || f.typeMeta == nil {
		return f
	}
	for _, fieldName := range fieldNames {
		structField := f.typeMeta.EnsureFieldByName(fieldName)
		if f.HasByIndex(structField.Index) {
			continue
		}
		f = f.definedOrEmptyCopy()
		f.fields[structField.Index] = newStructField(structField)
	}
	return f
}

// Remove deletes the field at the specified index
func (f Struct) Remove(fieldIndices ...int) Struct {
	if len(fieldIndices) == 0 {
		return f
	}
	if f.typeMeta == nil {
		panic("Cannot remove from invalid fragment")
	}
	ensuredFields := false
	for _, fieldIndex := range fieldIndices {
		if f.HasByIndex(fieldIndex) {
			if !ensuredFields {
				f = f.definedCopy()
				ensuredFields = true
			}
			delete(f.fields, fieldIndex)
		}
	}
	return f
}

// RemoveByName removes fields from the fragment
func (f Struct) RemoveByName(fieldNames ...string) Struct {
	if len(fieldNames) == 0 || f.typeMeta == nil {
		return f
	}
	ensuredFields := false
	for _, fieldName := range fieldNames {
		fieldIndex := f.typeMeta.EnsureFieldByName(fieldName).Index
		if f.HasByIndex(fieldIndex) {
			if !ensuredFields {
				f = f.definedCopy()
				ensuredFields = true
			}
			delete(f.fields, fieldIndex)
		}
	}
	return f
}

// Set sets the fragment of a field at the specified index of the fragment (and adds the field if it has not already been added)
func (f Struct) Set(fieldIndex int, fieldFragment interface{}) Struct {
	if f.typeMeta == nil {
		panic("Cannot set to invalid fragment")
	}
	structField := f.typeMeta.EnsureField(fieldIndex)
	parsedFieldFragment, err := ParseStruct(structField.TypeMeta, fieldFragment)
	if err != nil {
		panic("Invalid fragment for field \"" + f.typeMeta.Name() + "." + structField.String() + "\": " + err.Error())
	}
	f = f.definedOrEmptyCopy()
	if field, ok := f.fields[fieldIndex]; !ok {
		f.fields[fieldIndex] = StructField{StructField: structField, Fragment: parsedFieldFragment}
	} else {
		field.Fragment = parsedFieldFragment
		f.fields[fieldIndex] = field
	}
	return f
}

// SetByName sets the fragment of a field of the fragment (and adds the field if it has not already been added)
func (f Struct) SetByName(fieldName string, fieldFragment interface{}) Struct {
	if f.typeMeta == nil {
		return f
	}
	structField := f.typeMeta.EnsureFieldByName(fieldName)
	return f.Set(structField.Index, fieldFragment)
}

// Has parses the specified fragment and checks if every specified field is included in f.
func (f Struct) Has(v ...interface{}) bool {
	if f.typeMeta == nil {
		return true
	}
	hf, err := ParseStruct(f.typeMeta, v...)
	if err != nil {
		return false
	}
	if hf.IsUndefined() {
		return false
	}
	_, foundMissing := hf.FindField(func(field StructField) bool {
		if field.Fragment.IsUndefined() {
			return !f.HasByIndex(field.Index)
		} else if !f.HasByIndex(field.Index) {
			return true
		} else {
			return !f.FieldFragment(field.Index).Has(field.Fragment)
		}
	})
	return !foundMissing
}

// Has returns whether a field at the specified index is included in the fragment
func (f Struct) HasByIndex(fieldIndex int) bool {
	if f.typeMeta == nil {
		return true
	}
	structField := f.typeMeta.Field(fieldIndex)
	if structField == nil {
		return false
	} else if f.IsUndefined() {
		return !typeHasIncludeDefaults(f.typeMeta) || includeDefault(*structField)
	}
	_, has := f.fields[fieldIndex]
	return has
}

// HasByName returns whether the field with the specified name is included in the fragment.
// If the fragment is undefined, `Has` always returns true.
// If a struct field with the specified name does not exist in the struct, false is returned.
func (f Struct) HasByName(fieldName string) bool {
	if f.typeMeta == nil {
		return true
	}
	fieldIndex := f.typeMeta.FieldIndexByName(fieldName)
	if fieldIndex == -1 {
		return false
	}
	return f.HasByIndex(fieldIndex)
}

// AssignToField assigns the fragment of a field at the specified index of the fragment (and adds the field if it has not already been added)
func (f Struct) AssignToField(fieldIndex int, fieldFragment interface{}) Struct {
	if f.typeMeta == nil {
		panic("Cannot assign to invalid fragment")
	}
	structField := f.typeMeta.EnsureField(fieldIndex)
	parsedFieldFragment, err := ParseStruct(structField.TypeMeta, fieldFragment)
	if err != nil {
		panic("Invalid fragment for field \"" + f.typeMeta.Name() + "." + structField.String() + "\": " + err.Error())
	}
	f = f.definedOrEmptyCopy()
	if field, ok := f.fields[fieldIndex]; !ok {
		f.fields[fieldIndex] = StructField{structField, parsedFieldFragment}
	} else if field.Fragment.IsUndefined() {
		field.Fragment = parsedFieldFragment
		f.fields[fieldIndex] = field
	} else {
		field.Fragment = f.fields[fieldIndex].Fragment.assign(parsedFieldFragment)
		f.fields[fieldIndex] = field
	}
	return f
}

// // AddPath adds a field at the specified path
// func (f Struct) AddPath(fieldPath ...string) Struct {
// 	if f.HasPath(fieldPath...) {
// 		return f
// 	}
// 	field := f.Field(fieldPath[0])
// 	if field == nil {
// 		f.Add(fieldPath[0])
// 		field = f.Field(fieldPath[0])
// 	}
// 	if len(fieldPath) == 1 {
// 		return f
// 	}
// 	field.StructField.AddPath(fieldPath[1:]...)
// 	return f
// }

// // HasPath returns whether a field is defined at the specified path
// func (f Struct) HasPath(fieldPath ...string) bool {
// 	if f == nil {
// 		return true
// 	}
// 	field := f.Field(fieldPath[0])
// 	if field == nil {
// 		return false
// 	} else if len(fieldPath) == 1 {
// 		return true
// 	}
// 	return field.StructField.HasPath(fieldPath[1:]...)
// }

// // AddIndexPath adds a field at the specified path
// func (f Struct) AddIndexPath(fieldIndexPath ...int) Struct {
// 	pathLen := len(fieldIndexPath)
// 	if pathLen == 0 {
// 		return f
// 	}
// 	if f.HasIndexPath(fieldIndexPath...) {
// 		return f
// 	}
// 	field := f.Field(fieldIndexPath[0])
// 	if field == nil {
// 		f.AddIndex(fieldIndexPath[0])
// 		field = f.Field(fieldIndexPath[0])
// 	}
// 	if pathLen == 1 {
// 		return f
// 	}
// 	field.StructField = field.StructField.Ensure(field.Meta.typeMeta).AddIndexPath(fieldIndexPath[1:]...)
// 	return f
// }

// // AssignAtPath sets a fragment at the specified path
// func (f Struct) AssignAtPath(fragment Struct, fieldIndexPath ...int) Struct {
// 	if f == nil {
// 		return nil
// 	}
// 	field := f.Field(fieldIndexPath[0])
// 	if field == nil {
// 		f.AddIndex(fieldIndexPath[0])
// 		field = f.Field(fieldIndexPath[0])
// 	}
// 	if len(fieldIndexPath) == 1 {
// 		if field.StructField == nil {
// 			field.StructField = fragment
// 		} else {
// 			field.StructField.Assign(fragment)
// 		}
// 		return f
// 	}
// 	field.StructField.AssignAtPath(fragment, fieldIndexPath[1:]...)
// 	return f
// }

// // HasIndexPath returns whether a field is defined at the specified path
// func (f Struct) HasIndexPath(fieldIndexPath ...int) bool {
// 	if f == nil {
// 		return true
// 	}
// 	field := f.Field(fieldIndexPath[0])
// 	if field == nil {
// 		return false
// 	} else if len(fieldIndexPath) == 1 {
// 		return true
// 	}
// 	return field.StructField.HasIndexPath(fieldIndexPath[1:]...)
// }

// FieldByName returns a field of the fragment. Note that the field may be undefined even though `HasIndex` returns true.
// If a field should be ensured to be non-nil when `Has` returns true, `FieldOrMeta` can be used, which can
// return a field that the fragment do not have reference to (with only field meta defined).
func (f Struct) FieldByName(fieldName string) StructField {
	if f.typeMeta == nil {
		return StructField{}
	}
	return f.Field(f.typeMeta.EnsureFieldByName(fieldName).Index)
}

// FieldFragmentByName returns the fragment of a field of the fragment.
func (f Struct) FieldFragmentByName(fieldName string) Struct {
	field := f.FieldByName(fieldName)
	return field.Fragment
}

// Field returns a field of the fragment result. Note that the field may be undefined even though `HasIndex` returns true.
// If a field should be ensured to be non-nil when `HasIndex` returns true, `FieldOrMetaAt` can be used, which can
// return a field that the fragment do not have reference to (with only field meta defined).
func (f Struct) Field(fieldIndex int) StructField {
	if f.typeMeta == nil {
		return StructField{}
	}
	field, ok := f.fields[fieldIndex]
	if !ok {
		structField := f.typeMeta.EnsureField(fieldIndex)
		return StructField{StructField: structField, Fragment: Struct{typeMeta: typemeta.StructOf(structField.TypeMeta)}}
	}
	return field
}

// // FieldOrMeta returns a field of the fragment.
// func (f Struct) FieldOrMeta(fieldName string) StructField {
// 	if f == nil {
// 		return nil
// 	}
// 	return f.FieldOrMetaAt(f.typeMeta.EnsureFieldByName(fieldName).Index)
// }

// // FieldOrTypeMeta returns a field of the fragment result. If it is not defined, an unreferenced field with type meta is returned.
// func (f Struct) FieldOrTypeMeta(fieldIndex int) StructField {
// 	if f == nil {
// 		panic("attempted to get field of nil struct fragment")
// 	}
// 	if f.fields == nil || f.fields[fieldIndex] == nil {
// 		structField := f.typeMeta.EnsureField(fieldIndex)
// 		if typeHasIncludeDefaults(f.typeMeta) && !includeDefault(structField) {
// 			return nil
// 		}
// 		return &StructField{StructField: structField}
// 	}
// 	return f.fields[fieldIndex]
// }

// FieldFragment returns the fragment for the field at the specified index.
// Of course, this should only be used if the field is included in the fragment
func (f Struct) FieldFragment(fieldIndex int) Struct {
	return f.Field(fieldIndex).Fragment
}

// IterateFields iterates fields of the fragment
func (f Struct) IterateFields(iteratee func(field StructField)) {
	f.iterateFields(iteratee, false)
}

// IterateExplicitFields iterates fields of the fragment
func (f Struct) IterateExplicitFields(iteratee func(field StructField)) {
	f.iterateFields(iteratee, true)
}

func (f Struct) iterateFields(iteratee func(field StructField), doNotOmitDefault bool) {
	f.findField(func(field StructField) bool {
		iteratee(field)
		return false
	}, doNotOmitDefault)
}

// FindField iterates fields of the fragment
func (f Struct) FindField(iteratee func(field StructField) bool) (StructField, bool) {
	return f.findField(iteratee, false)
}

func (f Struct) findField(iteratee func(field StructField) bool, doNotOmitDefault bool) (StructField, bool) {
	if f.typeMeta == nil {
		return StructField{}, false
	}
	if f.IsUndefined() {
		hasIncludeDefaults := typeHasIncludeDefaults(f.typeMeta)
		var foundField StructField
		f.typeMeta.FindField(func(structField typemeta.StructField) bool {
			if !doNotOmitDefault && hasIncludeDefaults && !includeDefault(structField) {
				return false
			}
			fragmentField := newStructField(structField)
			if iteratee(fragmentField) {
				foundField = fragmentField
				return true
			}
			return false
		})
		if foundField.TypeMeta == nil {
			return foundField, false
		}
		return foundField, true
	}
	for _, fieldIndex := range f.fieldIndices() {
		field := f.fields[fieldIndex]
		if iteratee(field) {
			return field, true
		}
	}
	return StructField{}, false
}

func (f Struct) fieldIndices() []int {
	indices := []int{}
	if f.fields != nil {
		for index := range f.fields {
			indices = append(indices, index)
		}
		sort.Ints(indices)
	}
	return indices
}

// Omit creates a copy of the fragment and omits the specified fragment. It panics if the specified
// fragment is invalid or not assignable to the fragment.
func (f Struct) Omit(v ...interface{}) Struct {
	of, err := ParseStruct(f.typeMeta, v...)
	if err != nil {
		panic(err)
	}
	return f.omit(of)
}

func (f Struct) omitv(v ...interface{}) Fragment {
	of, err := ParseStruct(f.typeMeta, v...)
	if err != nil {
		panic(err)
	}
	return f.omit(of)
}

// Omit creates a copy of the fragment and omits the specified fragment. It panics if the specified
// fragment is invalid or not assignable to the fragment.
func (f Struct) omit(of Struct) Struct {
	if of.IsUndefinedOrEmpty() {
		return f
	}
	nf := f.definedCopy()
	remove := []int{}
	of.IterateFields(func(field StructField) {
		if field.Fragment.IsUndefined() {
			remove = append(remove, field.Index)
		} else {
			field.Fragment.omitv(field.Fragment)
		}
	})
	return nf.Remove(remove...)
}

// Pick returns a copy of the fragment with the specified fields picked
func (f Struct) Pick(v ...interface{}) Struct {
	pf, err := ParseStruct(f.typeMeta, v...)
	if err != nil {
		panic(err)
	}
	return f.pick(pf)
}

func (f Struct) pickv(v ...interface{}) Fragment {
	pf, err := ParseStruct(f.typeMeta, v...)
	if err != nil {
		panic(err)
	}
	return f.pick(pf)
}

// Pick returns a copy of the fragment with the specified fields picked
func (f Struct) pick(pf Struct) Struct {
	f = f.EnsureDefined()
	remove := []int{}
	pf.IterateFields(func(field StructField) {
		if !f.HasByIndex(field.Index) {
			// the field is not included in the current fragment, so remove it
			remove = append(remove, field.Index)
		} else {
			currentField := f.Field(field.Index)
			if field.Fragment.IsUndefined() {
				// if the picked fragment has a undefined fragment for the field,
				// set the field fragment to the current fragment of the field
				// currentField.Fragment = currentField.Fragment
			} else if currentField.Fragment.IsUndefined() {
				// if the current fragment is undefined, use the picked fragment
				currentField.Fragment = field.Fragment
			} else {
				// both current and picked fragments are non-nil, so pick from the current fragment
				currentField.Fragment = currentField.Fragment.pick(field.Fragment)
			}
		}
	})
	pf.Remove(remove...)
	return pf
}

// Assign assigns fragments deeply to the fragment
func (f Struct) Assign(v ...interface{}) Struct {
	af, err := ParseStruct(f.typeMeta, v...)
	if err != nil {
		panic(err)
	}
	return f.assign(af)
}

// Assign assigns fragments deeply to the fragment
func (f Struct) AssignStruct(af Struct) Struct {
	if af.typeMeta.Type() != f.typeMeta.Type() {
		panic("Expected fragment with type " + f.typeMeta.String() + " but received type " + af.typeMeta.String())
	}
	return f.assign(af)
}

func (f Struct) assignv(v ...interface{}) Fragment {
	af, err := ParseStruct(f.typeMeta, v...)
	if err != nil {
		panic(err)
	}
	return f.assign(af)
}

func (f Struct) assign(af Struct) Struct {
	add := []int{}
	af.IterateFields(func(field StructField) {
		if field.Fragment.IsUndefined() {
			// assigned fragment is undefined for field, so just make sure it is included in the fragment
			add = append(add, field.Index)
		} else if !f.HasByIndex(field.Index) {
			// fragment does not include field, so set the field with the assigned fragment
			f = f.Set(field.Index, field.Fragment)
		} else {
			// current field exists and assigned fragment is non-nil, so assign to it
			currentField := f.Field(field.Index)
			if currentField.Fragment.IsUndefined() {
				f = f.Set(field.Index, field.Fragment)
			} else {
				f = f.Set(field.Index, currentField.Fragment.assignv(field.Fragment))
			}
		}
	})
	return f.Add(add...)
}

// ToType returns a new fragment for the specified type with all matching parts of the current fragment.
func (f Struct) TryToType(t interface{}) (Struct, error) {
	if f.IsUndefined() {
		return NewStruct(t), nil
	}
	structTypeMeta := typemeta.StructOf(typemeta.Get(t))
	if structTypeMeta.Type() == f.TypeMeta().Type() {
		return f, nil
	}
	return ParseStruct(structTypeMeta, f.ToUnstructured())
}

// ToType returns a new fragment for the specified type with all matching parts of the current fragment. It panics if it fails.
func (f Struct) ToType(t interface{}) Struct {
	nf, err := f.TryToType(t)
	if err != nil {
		panic(err)
	}
	return nf
}

// FieldsLen returns the amount the fields
func (f Struct) FieldsLen() int {
	if f.fields != nil {
		return len(f.fields)
	}
	fieldsLen := 0
	f.IterateFields(func(field StructField) {
		fieldsLen++
	})
	return fieldsLen
}

// ToUnstructured returns an unstructured fragment based on the fragment
func (f Struct) ToUnstructured() Unstructured {
	json, err := ParseUnstructured(f.JSONExpr())
	if err != nil {
		panic("failed converting struct fragment to JSON fragment: " + err.Error())
	}
	return json
}

// EnsureValid ensures that the fragment is valid
func (f Struct) EnsureValid(structTypeMeta *typemeta.Struct) Struct {
	if f.typeMeta == nil {
		f.typeMeta = structTypeMeta
	} else if f.typeMeta.Type() != structTypeMeta.Type() {
		panic("Expected fragment with type " + structTypeMeta.String() + " but found type " + f.typeMeta.String())
	}
	return f
}

// EnsureDefined ensures that the fragment is non-nil
func (f Struct) EnsureDefined() Struct {
	if f.typeMeta == nil {
		panic("Attempted to ensure defined fragment with missing type meta")
	}
	if f.IsUndefined() {
		newFields := map[int]StructField{}
		f.IterateFields(func(field StructField) {
			newFields[field.Index] = field
		})
		f.fields = newFields
	}
	return f
}

func (f Struct) definedOrEmptyCopy() Struct {
	newFields := map[int]StructField{}
	if f.fields != nil {
		for index, field := range f.fields {
			newFields[index] = field
		}
	}
	f.fields = newFields
	return f
}

func (f Struct) definedCopy() Struct {
	if f.IsUndefined() {
		f = f.EnsureDefined()
	} else {
		newFields := map[int]StructField{}
		for index, field := range f.fields {
			newFields[index] = field
		}
		f.fields = newFields
	}
	return f
}

func (f Struct) Copy() Struct {
	if !f.IsUndefined() {
		newFields := map[int]StructField{}
		for index, field := range f.fields {
			newFields[index] = field
		}
		f.fields = newFields
	}
	return f
}

// Copy returns a copy of the fragment
func (f *Struct) CopyPtr() *Struct {
	if f == nil {
		return nil
	}
	cf := f.Copy()
	return &cf
}

// func (f Struct) copy() Struct {
// 	if f.fields == nil {
// 		return f
// 	} else {
// 		newFields := map[int]StructField{}
// 		for index, field := range f.fields {
// 			newFields[index] = field
// 		}
// 		f.fields = newFields
// 	}
// 	return f
// }

// Clear sets the fragment to empty
func (f Struct) Clear() Struct {
	f.fields = map[int]StructField{}
	return f
}

// IsUndefined returns whether the fragment doesn't have any specified fields
func (f Struct) IsUndefined() bool {
	return f.fields == nil || f.typeMeta == nil
}

// IsValid returns whether the fragment has a type defined
func (f Struct) IsValid() bool {
	return f.typeMeta != nil
}

// // IsNilOrUndefined returns whether the fragment is nil or doesn't have any specified fields
// func (f *Struct) IsNilOrUndefined() bool {
// 	return f == nil || f.fields == nil
// }

// // IsNilOrUndefined returns whether the fragment is nil or doesn't have any specified fields
// func (f *Struct) IsNilOrUndefinedOrEmpty() bool {
// 	return f == nil || f.fields == nil || len(f.fields) == 0
// }

// IsEmpty returns whether the fragment have explicitly specified no fields
func (f Struct) IsEmpty() bool {
	return !f.IsUndefined() && len(f.fields) == 0
}

// IsUndefinedOrEmpty returns whether the fragment is undefined or don't have any specified fields
func (f Struct) IsUndefinedOrEmpty() bool {
	return f.IsUndefined() || len(f.fields) == 0
}

// TypeMeta returns the type meta of the fragment (always *typemeta.Struct).
// It returns nil if the fragment is invalid.
func (f Struct) TypeMeta() *typemeta.Struct {
	return f.typeMeta
}

// Primitive returns false for struct fragments
func (f Struct) Primitive() bool {
	return false
}

func (f Struct) String() string {
	if !f.IsValid() {
		return "Fragment(<invalid>)"
	}
	return f.typeMeta.Name() + "Fragment(" + f.Expr() + ")"
}

// Expr returns a fragment expression
func (f Struct) Expr() string {
	return f.expr()
}

// JSONExpr returns a JSON fragment expression
func (f Struct) JSONExpr() string {
	expr := ""
	f.IterateFields(func(field StructField) {
		fieldExpr := field.JSONExpr()
		if fieldExpr == "" {
			return
		}
		if expr != "" {
			expr += ", " + fieldExpr
		} else {
			expr = fieldExpr
		}
	})
	if expr == "" {
		return ""
	}
	return "{ " + expr + " }"
}

// Expr returns a fragment expression
func (sf StructField) Expr() string {
	fragmentExpr := sf.Fragment.Expr()
	if fragmentExpr != "" {
		return sf.Name + " " + fragmentExpr
	}
	return sf.expr()
}

func (sf StructField) JSONExpr() string {
	if sf.JSONName == "" {
		return ""
	} else if sf.Fragment.IsUndefined() {
		return sf.JSONName
	}
	fragmentJSONExpr := sf.Fragment.JSONExpr()
	if fragmentJSONExpr != "" {
		return sf.JSONName + " " + fragmentJSONExpr
	}
	return sf.JSONName
}

func (f Struct) expr() string {
	expr := ""
	f.IterateFields(func(field StructField) {
		fieldExpr := field.expr()
		if expr != "" {
			expr += ", " + fieldExpr
		} else {
			expr += fieldExpr
		}
	})
	if expr == "" {
		return ""
	}
	return "{ " + expr + " }"
}

func (sf StructField) expr() string {
	if sf.Fragment.IsUndefined() {
		return sf.Name
	}
	fragmentExpr := sf.Fragment.Expr()
	if fragmentExpr != "" {
		return sf.Name + " " + fragmentExpr
	}
	return sf.Name
}

// UndefinedStruct is an undefined struct fragment
var UndefinedStruct Struct
