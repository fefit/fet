package types

// NamespaceFn for variable
type NamespaceFn func(name string) (bool, string)

// Indexs struct
type Indexs struct {
	StartIndex int
	EndIndex   int
}

// Mode of parse type
type Mode int

// Smarty for Mode
const (
	Gofet Mode = 1 << iota
	Smarty
	AnyMode = Gofet | Smarty
)
