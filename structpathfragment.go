package fragment

// FragmentAt returns the fragment of the specified fragment at the struct path
func (sp StructPath) FragmentAt(fragment Struct) Struct {
	currentFragment := fragment
	for _, fieldIndex := range sp.FieldIndices() {
		fragment = fragment.FieldFragment(fieldIndex)
	}
	return currentFragment
}
