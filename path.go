package fragment

import "github.com/ludvigalden/go-typemeta"

// Path is the interface for a fragment.
type Path interface {
	FieldNames() []string
	JSONFieldNames() []string
	ToFragment() Fragment
	TailFragment() Fragment
	String() string
	Expr() string
	JSONExpr() string
}

// TypedPath is the interface for a fragment.
type TypedPath interface {
	Path
	FieldIndices() []int
	TypeMeta() typemeta.TypeMeta
}
