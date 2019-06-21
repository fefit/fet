package types

// NamespaceFn for variable
type NamespaceFn func(name string) (bool, string)

// Indexs struct
type Indexs struct {
	StartIndex int
	EndIndex   int
}
