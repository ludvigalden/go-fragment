package fragment

// Fragment is the interface for a fragment.
type Fragment interface {
	pickv(...interface{}) Fragment
	omitv(...interface{}) Fragment
	assignv(...interface{}) Fragment
	Has(...interface{}) bool
	HasByName(fieldName string) bool
	IsUndefined() bool
	IsEmpty() bool
	IsUndefinedOrEmpty() bool
	String() string
	Expr() string
	JSONExpr() string
}
