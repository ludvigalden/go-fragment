package fragment

import "strings"

// UnstructuredPath is an interface path
type UnstructuredPath struct {
	fieldNames   []string
	tailFragment Fragment
}

var _ Path = (&UnstructuredPath{})

// TailFragment returns the tail fragment of the path
func (ip UnstructuredPath) TailFragment() Fragment {
	return ip.tailFragment
}

// FieldNames returns the head fragment
func (ip UnstructuredPath) FieldNames() []string {
	if ip.fieldNames == nil {
		return nil
	}
	return ip.fieldNames
}

// JSONFieldNames returns the head fragment
func (ip UnstructuredPath) JSONFieldNames() []string {
	return ip.FieldNames()
}

// ToFragment returns a fragment representing the path
func (ip UnstructuredPath) ToFragment() Fragment {
	fieldNames := ip.FieldNames()
	if fieldNames == nil {
		return nil
	}
	root := NewUnstructured()
	current := root
	maxIndex := len(fieldNames) - 1
	for i, key := range fieldNames {
		if i == maxIndex {
			if ip.tailFragment != nil && !ip.tailFragment.IsUndefined() {
				current = current.Set(key, ip.tailFragment)
			} else {
				current = current.Add(key)
			}
		} else {
			current = current.Set(key, Unstructured{})
			current = current.Field(key).(Unstructured)
		}
	}
	return root
}

// Expr returns an expression for the interface path
func (ip UnstructuredPath) Expr() string {
	fieldNames := ip.FieldNames()
	if fieldNames == nil {
		return ""
	}
	expr := strings.Join(fieldNames, ".")
	if ip.tailFragment != nil {
		tailFragmentExpr := ip.tailFragment.Expr()
		if tailFragmentExpr != "" {
			expr += " " + tailFragmentExpr
		}
	}
	return expr
}

// JSONExpr returns a JSON expression for the interface path
func (ip UnstructuredPath) JSONExpr() string {
	jsonFieldNames := ip.JSONFieldNames()
	if jsonFieldNames == nil {
		return ""
	}
	expr := strings.Join(jsonFieldNames, ".")
	if ip.tailFragment != nil {
		tailFragmentJSONExpr := ip.tailFragment.JSONExpr()
		if tailFragmentJSONExpr != "" {
			expr += " " + tailFragmentJSONExpr
		}
	}
	return expr
}

func (ip UnstructuredPath) String() string {
	return "Path(" + ip.Expr() + ")"
}
