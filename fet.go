package fet

type Config struct {
	LeftDelimiter  []byte
	RightDelimiter []byte
}

type Engine interface {
	Config(config Config)
}

func NewEngine() {

}

type BlockFeature struct {
	MustLast bool
}

type Block struct {
	Name     []byte
	Features []BlockFeature
}
