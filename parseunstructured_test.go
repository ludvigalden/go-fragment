package fragment

import (
	"fmt"
	"strconv"
	"testing"
)

func TestParseUnstructured(t *testing.T) {
	t.Run("parses comma layers", func(t *testing.T) {
		matches := []struct {
			e  string
			el int
		}{{"{}", 1}, {"a, b, c { d, f, g }, h", 4}, {"ab,c", 2}, {"ab,c,d{f,g}", 3}}
		for _, match := range matches {
			r := []rune(match.e)
			rp, err := splitFragmentByComma(r, len(r))
			if err != nil {
				t.Error("did not expect `splitFragmentByComma` to return error for valid fragment: " + err.Error())
				return
			}
			if len(rp) != match.el {
				t.Error("expected `splitFragmentByComma` to return " + strconv.Itoa(match.el) + " parts for: " + match.e)
				return
			}
		}
	})
	t.Run("parses fragment layers", func(t *testing.T) {
		matches := []struct {
			e string
			p []string
		}{{"a {}", []string{"a ", "{}"}}, {"a, b, c, d { f, g }", []string{"a, b, c, d ", "{ f, g }"}}, {"ab,c", []string{"ab,c"}}, {"ab,c {}", []string{"ab,c ", "{}"}}, {"ab,c,d{f,g}, h {i,j}", []string{"ab,c,d", "{f,g}", ", h ", "{i,j}"}}}
		for _, match := range matches {
			r := []rune(match.e)
			rp, err := splitFragment(r, len(r))
			if err != nil {
				t.Error("did not expect `splitFragment` to return error for valid fragment: " + err.Error())
				return
			}
			if len(rp) != len(match.p) {
				t.Error("expected `splitFragment` to return " + strconv.Itoa(len(match.p)) + " parts for: " + match.e)
				return
			}
			for i, p := range match.p {
				if string(rp[i]) != p {
					t.Error("expected `splitFragment` to return part \"" + p + "\", but received \"" + string(rp[i]) + "\"")
				}
			}
		}
	})
	t.Run("parses expressions", func(t *testing.T) {
		fragment, err := ParseUnstructured("fieldA, fieldB, fieldC")
		if err != nil {
			t.Error("did not expect `ParseUnstructured` to return error for valid fragment: " + err.Error())
			return
		}
		if fragment.IsUndefined() {
			t.Error("expected fragment to not be undefined")
			return
		}
		if fragment.FieldsLen() != 3 {
			t.Error("expected `FieldsLen` to return `3`, but received `" + fmt.Sprint(fragment.FieldsLen()) + "`")
			return
		}
		fragment, err = ParseUnstructured("fieldA { fieldB { fieldC } }")
		if err != nil {
			t.Error("did not expect `ParseUnstructured` to return error for valid fragment: " + err.Error())
			return
		}
		if !fragment.HasByName("fieldA") {
			t.Error("expected `Has` to return true for included field in " + fragment.String())
			return
		}
		if !fragment.Field("fieldA").(Unstructured).HasByName("fieldB") {
			t.Error("expected `Has` to return true for included nested field")
			return
		}
		if !fragment.Field("fieldA").(Unstructured).Field("fieldB").(Unstructured).HasByName("fieldC") {
			t.Error("expected `Has` to return true for included nested field")
			return
		}
		if fragment.Field("fieldA").(Unstructured).Field("fieldB").(Unstructured).HasByName("fieldD") {
			t.Error("expected `Has` to return false for non-included nested field")
			return
		}

		fragment, err = ParseUnstructured("")
		if err != nil {
			t.Error("did not expect `ParseUnstructured` to return error for valid fragment: " + err.Error())
			return
		}
		if !fragment.IsUndefined() {
			t.Error("expected `IsUndefined` to return true for non-specified fragment")
			return
		}

		fragment, err = ParseUnstructured("{}")
		if err != nil {
			t.Error("did not expect `ParseUnstructured` to return error for valid fragment: " + err.Error())
			return
		}
		if fragment.IsUndefined() {
			t.Error("expected `IsUndefined` to return false for empty specified fragment")
			return
		}

		invalidFragments := []string{"fieldA {", "{"}
		for _, invalidFragment := range invalidFragments {
			fragment, err = ParseUnstructured(invalidFragment)
			if err == nil {
				t.Error("expected `ParseUnstructured` to return an error for invalid fragment \"" + invalidFragment + "\", but received " + fragment.String())
				return
			}
		}

		validFragments := []string{"{ fieldA, fieldB }", "fieldA, ", "field A B C", "fieldA,", "field, A, B, C", "fieldC.fieldD.fieldE.fieldF"}
		for _, validFragment := range validFragments {
			fragment, err = ParseUnstructured(validFragment)
			if err != nil {
				t.Error("expected `ParseUnstructured` to not return an error for valid fragment \"" + validFragment + "\": " + err.Error())
				return
			}
		}

		matches := [][2]string{{"fieldA,fieldB{fieldC}", "{ fieldA, fieldB { fieldC } }"}}
		for _, match := range matches {
			fragment, err = ParseUnstructured(match[0])
			if err != nil {
				t.Error("expected `ParseUnstructured` to not return an error for valid fragment \"" + match[0] + "\": " + err.Error())
				return
			}
			if match[1] != fragment.Expr() {
				t.Log("expected `Expr` to return \"" + match[1] + "\" for parsed \"" + match[0] + "\", but received \"" + fragment.Expr() + "\"")
			}
		}
	})
}
