package fragment

import (
	"sync"
	"testing"
	"time"
)

func TestStruct(t *testing.T) {
	type StructA struct {
		Name string                 `json:"name"`
		Age  uint32                 `json:"age"`
		Info map[string]interface{} `json:"info"`
	}
	t.Run("check included immediate fields", func(t *testing.T) {
		fragment := NewStruct(StructA{})
		if !fragment.HasByName("Name") || !fragment.HasByName("Age") || !fragment.HasByName("Info") {
			t.Error("Has returns false for included fields")
			return
		}
		fragment = fragment.AddByName("Name")
		if !fragment.HasByName("Name") {
			t.Error("Has returns false for included fields")
			return
		}
		if fragment.HasByName("Age") {
			t.Error("Has returns true for excluded fields")
			return
		}
	})
	t.Run("assigns", func(t *testing.T) {
		fragment1 := NewStruct(StructA{}).AddByName("Name")
		fragment2 := NewStruct(StructA{}).AddByName("Info")
		fragment := NewStruct(StructA{})
		fragment = fragment.Assign(fragment1, fragment2)
		if fragment.HasByName("Age") {
			t.Error("Has returns true for excluded fields")
			return
		}
		if !fragment.HasByName("Name") || !fragment.HasByName("Info") {
			t.Error("Has returns false for included fields")
			return
		}
	})
	t.Run("copies used in multiple goroutines", func(t *testing.T) {
		root := NewEmptyStruct(StructA{})
		wg := sync.WaitGroup{}
		wg.Add(3)
		copy1 := root
		go func() {
			defer wg.Done()
			if copy1.HasByName("Name") {
				t.Error("Has returns true for excluded fields (copy1)")
				return
			}
			if copy1.HasByName("Age") {
				t.Error("Has returns true for excluded fields (copy1)")
				return
			}
			if copy1.HasByName("Info") {
				t.Error("Has returns true for excluded fields (copy1)")
				return
			}
			copy1 = copy1.Assign("Age, Info")
			if copy1.HasByName("Name") {
				t.Error("Has returns true for excluded fields (copy1)")
				return
			}
			if !copy1.HasByName("Age") {
				t.Error("Has returns false for included fields (copy1)")
				return
			}
			if !copy1.HasByName("Info") {
				t.Error("Has returns false for included fields (copy1)")
				return
			}
		}()
		copy2 := root
		go func() {
			defer wg.Done()
			if copy2.HasByName("Name") {
				t.Error("Has returns true for excluded fields (copy2)")
				return
			}
			if copy2.HasByName("Age") {
				t.Error("Has returns true for excluded fields (copy2)")
				return
			}
			if copy2.HasByName("Info") {
				t.Error("Has returns true for excluded fields (copy2)")
				return
			}
			copy2 = copy2.Assign("Name, Age")
			if !copy2.HasByName("Name") {
				t.Error("Has returns false for included fields (copy2)")
				return
			}
			if !copy2.HasByName("Age") {
				t.Error("Has returns false for included fields (copy2)")
				return
			}
			if copy2.HasByName("Info") {
				t.Error("Has returns true for excluded fields (copy2)")
				return
			}
		}()
		copy3 := root
		go func() {
			defer wg.Done()
			if copy3.HasByName("Name") {
				t.Error("Has returns true for excluded fields (copy3)")
				return
			}
			if copy3.HasByName("Age") {
				t.Error("Has returns true for excluded fields (copy3)")
				return
			}
			if copy3.HasByName("Info") {
				t.Error("Has returns true for excluded fields (copy3)")
				return
			}
			copy3 = copy3.Assign("Name, Info")
			if !copy3.HasByName("Name") {
				t.Error("Has returns false for included fields (copy3)")
				return
			}
			if copy3.HasByName("Age") {
				t.Error("Has returns true for excluded fields (copy3)")
				return
			}
			if !copy3.HasByName("Info") {
				t.Error("Has returns false for included fields (copy3)")
				return
			}
		}()
		wg.Wait()
	})
	type StructB struct {
		StructA StructA   `json:"a"`
		Date    time.Time `json:"time" fragment:"includedefault"`
	}
	t.Run("includedefault checks", func(t *testing.T) {
		fragment := NewStruct(StructB{})
		if fragment.HasByName("StructA") {
			t.Error("Has returns true for field excluded by default")
			return
		}
		if !fragment.HasByName("Date") {
			t.Error("Has returns false for field included by default")
			return
		}
		fragment = fragment.AddByName("StructA")
		if !fragment.HasByName("StructA") {
			t.Error("Has returns false for explicitly included field")
			return
		}
		if fragment.HasByName("Date") {
			t.Error("Has returns true explicitly excluded field")
			return
		}
		fragment = fragment.RemoveByName("StructA")
		if fragment.HasByName("StructA") {
			t.Error("Has returns true explicitly excluded field")
			return
		}
		if fragment.HasByName("Date") {
			t.Error("Has returns true explicitly excluded field")
			return
		}
	})
}
