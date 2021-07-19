package fragment

import (
	"testing"
)

func TestUnstructured(t *testing.T) {
	t.Run("constructs with fields", func(t *testing.T) {
		fragment := NewUnstructured().Add("fieldA", "fieldB", "fieldD")
		if !fragment.HasByName("fieldA") || !fragment.HasByName("fieldB") || !fragment.HasByName("fieldD") {
			t.Error("expected `Has` to return true for included fields")
			return
		}
		if fragment.HasByName("fieldC") {
			t.Error("expected `Has` to return false for non-included field")
		}
	})
	t.Run("undefined behaviour", func(t *testing.T) {
		var fragment Unstructured
		if !fragment.HasByName("fieldB") {
			t.Error("expected `Has` to return true for all fields when fragment is undefined")
			return
		}
		if fragment.Expr() != "" {
			t.Error("expected `Expr` to return an empty string for undefined fragments")
			return
		}
		if fragment.FieldsLen() != 0 {
			t.Error("expected `FieldsLen` to return 0 for undefined fragments")
			return
		}

	})
	t.Run("adds and removes unstructured fields", func(t *testing.T) {
		fragment := NewUnstructured()
		if !fragment.IsUndefined() {
			t.Error("expected `IsUndefined` to return true when no fields are specified")
			return
		}
		fragment = fragment.Add("fieldA")
		if fragment.IsUndefined() {
			t.Error("expected `IsUndefined` to return false when one or more fields are specified")
			return
		}
		if !fragment.HasByName("fieldA") {
			t.Error("expected `Has` to return true for just added field")
			return
		}
		fragment = fragment.Remove("fieldA")
		if fragment.HasByName("fieldA") {
			t.Error("expected `Has` to return false for just removed field")
			return
		}
		if fragment.IsUndefined() {
			t.Error("expected `IsUndefined` to return false when all fields are removed")
			return
		}
		if fragment.HasByName("fieldB") {
			t.Error("expected `Has` to return false for non-included field")
			return
		}
	})
}
