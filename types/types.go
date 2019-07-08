package types

// NamespaceFn for variable
type NamespaceFn func(name string) (bool, string)

// Indexs struct
type Indexs struct {
	StartIndex int
	EndIndex   int
}

// FetConfig struct
type FetConfig struct {
	LeftDelimiter  string
	RightDelimiter string
	CommentSymbol  string
	TemplateDir    string
	CompileDir     string
	UcaseField     bool
	CompileOnline  bool
	Glob           bool
	AutoRoot       bool
	Ignores        []string
	Mode           Mode
}

// Mode of parse type
type Mode int

// Smarty for Mode
const (
	Gofet Mode = 1 << iota
	Smarty
	AnyMode = Gofet | Smarty
)
