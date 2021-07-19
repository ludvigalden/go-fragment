package fragment

import (
	"testing"
)

func TestStructPath(t *testing.T) {
	type StructA struct {
		Name string
	}
	type StructB struct {
		Age int
		A   []StructA
	}
	type StructC struct {
		B StructB
	}
	t.Run("constructs nil", func(t *testing.T) {
		sp := NewStructPath(StructC{})
		if !sp.ToStructFragment().IsEmpty() {
			t.Error("Expected struct fragment to be empty")
		}
		sp = NewStructPath(StructC{}, "B.Age")
		if sp.ToStructFragment().IsUndefined() {
			t.Error("Expected struct fragment to not be defined")
		}
		if len(sp.FieldNames()) != 2 {
			t.Error("Expected field names len to be 2")
		}
		if len(sp.FieldIndices()) != 2 {
			t.Error("Expected field indices len to be 2")
		}
		if sp.ToStructFragment().Expr() != "{ B { Age } }" {
			t.Error("Expected struct fragment expression to be `{ B { Age } }`, but received " + sp.ToStructFragment().Expr())
		}
		sp = NewStructPath(StructC{}, "B.A.Name")
		if len(sp.FieldNames()) != 3 {
			t.Error("Expected field names len to be 3")
		}
		if len(sp.FieldIndices()) != 3 {
			t.Error("Expected field indices len to be 3")
		}
		if sp.ToStructFragment().Expr() != "{ B { A { Name } } }" {
			t.Error("Expected struct fragment expression to be `{ B { A { Name } } }`, but received " + sp.ToStructFragment().Expr())
		}

	})
}
