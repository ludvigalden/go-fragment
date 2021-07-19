package fragment

// Unstructured is an interface for a fragment, not specific to any type
type Unstructured struct {
	fields map[string]Fragment
}

var _ Fragment = Unstructured{}

// IsUndefined returns whether the fragment is undefined or does not have any specified fields
func (f Unstructured) IsUndefined() bool {
	return f.fields == nil
}

// Add adds a set of fields
func (f Unstructured) Add(fields ...string) Unstructured {
	f = f.definedCopy()
	if len(f.fields) == 0 {
		f.fields = map[string]Fragment{}
		for _, field := range fields {
			f.fields[field] = nil
		}
	} else {
		for _, field := range fields {
			if f.HasByName(field) {
				continue
			}
			f.fields[field] = nil
		}
	}
	return f
}

// Remove removes a set of fields
func (f Unstructured) Remove(fields ...string) Unstructured {
	if len(fields) == 0 {
		return f
	}
	f = f.definedCopy()
	if len(f.fields) == 0 {
		return f
	}
	for _, field := range fields {
		if f.HasByName(field) {
			delete(f.fields, field)
		}
	}
	return f
}

// Set sets a field
func (f Unstructured) Set(fieldName string, fieldFragment Fragment) Unstructured {
	if fieldFragment == nil {
		panic("The fragment for field \"" + fieldName + "\" was attempted to be set to nil")
	}
	f = f.definedCopy()
	f.fields[fieldName] = fieldFragment
	return f
}

// Has parses the specified fragment and checks if every specified field is included in f.
func (f Unstructured) Has(v ...interface{}) bool {
	hf, err := ParseUnstructured(v)
	if f.fields == nil {
		return true
	}
	if err != nil {
		return false
	}
	if hf.IsUndefined() {
		return false
	}
	missing := hf.FindField(func(fieldName string, fieldFragment Fragment) bool {
		if fieldFragment.IsUndefined() {
			return !f.HasByName(fieldName)
		} else if !f.HasByName(fieldName) {
			return true
		} else {
			currentFieldFragment := f.Field(fieldName)
			if currentFieldFragment == nil || currentFieldFragment.IsUndefined() {
				return false
			}
			return currentFieldFragment.Has(fieldFragment)
		}
	})
	return missing == nil
}

// HasByName returns whether a field with the specified name is included in the fragment. If the fragment is undefined, true is also returned.
func (f Unstructured) HasByName(field string) bool {
	if f.IsUndefined() {
		return true
	}
	_, ok := f.fields[field]
	return ok
}

// HasPath returns whether a field is fueried
func (f Unstructured) HasPath(fieldPath ...string) bool {
	if f.IsUndefined() || len(fieldPath) == 0 {
		return true
	}
	if !f.HasByName(fieldPath[0]) {
		return false
	}
	return f.Field(fieldPath[0]) != nil
}

// HasAny returns whether any field is fueried
func (f Unstructured) HasAny() bool {
	return f.fields != nil && len(f.fields) > 0
}

// Field returns the fragment of a field of the fragment. It may return nil even though the field is included in the fragment.
func (f Unstructured) Field(field string) Fragment {
	if f.fields != nil {
		return f.fields[field]
	}
	return nil
}

// IterateFields iterates fields of the fragment
func (f Unstructured) IterateFields(iteratee func(fieldName string, fieldFragment Fragment)) {
	if f.fields != nil {
		for fieldName, fieldFragment := range f.fields {
			iteratee(fieldName, fieldFragment)
		}
	}
}

// FindField iterates fields of the fragment
func (f Unstructured) FindField(iteratee func(fieldName string, fieldFragment Fragment) bool) Fragment {
	if f.fields != nil {
		for fieldName, fieldFragment := range f.fields {
			if iteratee(fieldName, fieldFragment) {
				return fieldFragment
			}
		}
	}
	return nil
}

// Pick returns a copy of the fragment with the specified fields picked
func (f Unstructured) Pick(v ...interface{}) Unstructured {
	pf, err := ParseUnstructured(v)
	if err != nil {
		panic(err)
	}
	return f.pick(pf)
}

func (f Unstructured) pickv(v ...interface{}) Fragment {
	pf, err := ParseUnstructured(v)
	if err != nil {
		panic(err)
	}
	return f.pick(pf)
}

// Pick returns a copy of the fragment with the specified fields picked
func (f Unstructured) pick(pf Unstructured) Unstructured {
	remove := []string{}
	pf.IterateFields(func(fieldName string, pickFieldFragment Fragment) {
		if !f.Has(fieldName) {
			// the field is not included in the current fragment, so remove it
			remove = append(remove, fieldName)
		} else {
			currentFieldFragment := f.Field(fieldName)
			if pickFieldFragment == nil || pickFieldFragment.IsUndefined() {
				// if the picked fragment has a undefined fragment for the field,
				// use the current fragment of the field
			} else if currentFieldFragment == nil || currentFieldFragment.IsUndefined() {
				// if the current fragment is undefined, use the picked fragment
				pf = pf.Set(fieldName, pickFieldFragment)
			} else {
				// both current and picked fragments are non-nil, so pick from the current fragment
				newFieldFragment := currentFieldFragment.pickv(pickFieldFragment)
				if newFieldFragment.IsEmpty() {
					remove = append(remove, fieldName)
				} else {
					pf = pf.Set(fieldName, newFieldFragment)
				}
			}
		}
	})
	pf = pf.Remove(remove...)
	return pf
}

// Omit creates a copy of the fragment and omits the specified fragment. It panics if the specified
// fragment is invalid or not assignable to the fragment.
func (f Unstructured) Omit(v ...interface{}) Fragment {
	of, err := ParseUnstructured(v)
	if err != nil {
		panic(err)
	}
	return f.omit(of)
}

func (f Unstructured) omitv(v ...interface{}) Fragment {
	of, err := ParseUnstructured(v)
	if err != nil {
		panic(err)
	}
	return f.omit(of)
}

// Omit creates a copy of the fragment and omits the specified fragment. It panics if the specified
// fragment is invalid or not assignable to the fragment.
func (f Unstructured) omit(of Unstructured) Unstructured {
	if of.IsUndefinedOrEmpty() {
		return f
	}
	remove := []string{}
	of.IterateFields(func(fieldName string, omitFieldFragment Fragment) {
		if omitFieldFragment == nil || omitFieldFragment.IsUndefined() {
			remove = append(remove, fieldName)
		} else {
			currentFieldFragment := of.Field(fieldName)
			if currentFieldFragment == nil || currentFieldFragment.IsUndefined() {
				remove = append(remove, fieldName)
			} else {
				newFieldFragment := currentFieldFragment.omitv(omitFieldFragment)
				if newFieldFragment.IsEmpty() {
					remove = append(remove, fieldName)
				} else {
					f = f.Set(fieldName, newFieldFragment)
				}
			}
		}
	})
	return f.Remove(remove...)
}

// Assign assigns fragments deeply to the fragment
func (f Unstructured) Assign(fragments ...interface{}) Unstructured {
	return f.assignv(fragments...).(Unstructured)
}

func (f Unstructured) assignv(fragments ...interface{}) Fragment {
	for _, assignI := range fragments {
		assign, err := ParseUnstructured(assignI)
		if err != nil {
			panic("Attempted to assign invalid fragment, " + err.Error())
		}
		if assign.fields != nil {
			for fieldName, fieldFragment := range assign.fields {
				if f.fields == nil {
					f.fields = map[string]Fragment{fieldName: fieldFragment}
				} else if f.fields[fieldName] == nil {
					f.fields[fieldName] = fieldFragment
				} else {
					f.fields[fieldName] = NewUnstructured().Assign(f.fields[fieldName], fieldFragment)
				}
			}
		}
	}
	return f
}

// IsEmpty returns whether the fragment have explicitly specified no fields
func (f Unstructured) IsEmpty() bool {
	return !f.IsUndefined() && len(f.fields) == 0
}

// IsUndefinedOrEmpty returns whether the fragment is undefined or don't have any specified fields
func (f Unstructured) IsUndefinedOrEmpty() bool {
	return f.IsUndefined() || len(f.fields) == 0
}

// FieldsLen returns the amount the fields
func (f Unstructured) FieldsLen() int {
	if f.fields == nil {
		return 0
	}
	return len(f.fields)
}

// EnsureDefined ensures that the fragment is non-nil
func (f Unstructured) EnsureDefined() Unstructured {
	if f.fields == nil {
		f.fields = map[string]Fragment{}
	}
	return f
}

func (f Unstructured) definedCopy() Unstructured {
	newFields := map[string]Fragment{}
	if f.fields != nil {
		for fieldName, fieldFragment := range f.fields {
			newFields[fieldName] = fieldFragment
		}
	}
	f.fields = newFields
	return f
}

func (f Unstructured) String() string {
	str := f.Expr()
	if str == "" {
		return "Fragment(*)"
	}
	return "Fragment(" + str + ")"
}

// Primitive returns false for unstructured fragments
func (f Unstructured) Primitive() bool {
	return false
}

// Expr returns the fragment string for the fragment
func (f Unstructured) Expr() string {
	if f.IsUndefined() {
		return ""
	}
	expr := ""
	for fieldName, fieldFragment := range f.fields {
		if expr != "" {
			expr += ", "
		}
		if fieldFragment == nil {
			expr += fieldName
		} else if fieldExpr := fieldFragment.Expr(); fieldExpr != "" {
			expr += (fieldName + " " + fieldExpr)
		} else {
			expr += fieldName
		}
	}
	if expr == "" {
		return "{}"
	}
	return "{ " + expr + " }"
}

// JSONExpr returns the unstructured fragment expression, since it contains no JSON metadata.
func (f Unstructured) JSONExpr() string {
	return f.Expr()
}
