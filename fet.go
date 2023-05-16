package fet

import (
	"github.com/fefit/fet/lexer"
)

type Config struct {
	LeftDelimiter  []byte
	RightDelimiter []byte
}

type Engine interface {
	Config(config Config)
}

func NewEngine() {

}

type IParser interface {
	// if the parser allowed clear
	Clear() bool
}

type ICode interface {
}

type Property struct {
	Shorthand      bool
	Required       bool
	Prop           string
	ValueValidator func(token lexer.IToken) error
}

/**
 * built-in property parser
 */
type BuiltInPropParser struct {
	Props []Property
}

type BlockFeature struct {
	Once   bool   // only allowed appear once
	Last   bool   // only allowed appear at last feature
	Name   []byte // feature block name
	Alias  []byte // feature block alias name, if exists
	Parser IParser
}

/**
 *
 */
type Block struct {
	Name     []byte
	EndName  []byte
	Features []BlockFeature
	Parser   IParser
	Childs   []ICode
}

type Context struct {
	Scope []byte // scope

}

type OutputCode struct {
	RawBytes [][]byte
}
