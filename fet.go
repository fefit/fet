package fet

import (
	"github.com/fefit/fet/lexer"
)

type InScope = uint
type PathType = uint
type Bytes = []byte

const (
	LocalScope InScope = iota
	ParentScope
	GlobalScope
)

const (
	AbsPath  PathType = iota // absolute path
	RelaPath                 // relative path
	BasePath                 // path base on
)

type TplPath struct {
	Type    PathType
	Path    []Bytes
	AbsPath []Bytes
}

type Config struct {
	LeftDelimiter  Bytes
	RightDelimiter Bytes
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

type LinkedNode struct {
	Parent     ICode // parent node
	Prev       ICode // prev node
	Next       ICode // next node
	FirstChild ICode // first child
}

/**
 * Property
 */

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
	Once   bool  // only allowed appear once
	Last   bool  // only allowed appear at last feature
	Name   Bytes // feature block name
	Alias  Bytes // feature block alias name, if exists
	Parser IParser
}

/**
 *
 */
type Block struct {
	Name     Bytes
	EndName  Bytes
	Features []BlockFeature
	Parser   IParser
	Node     *LinkedNode
}

type Context struct {
	Scope Bytes // scope

}

/**
 * Raw Html
 */
type RawHtml struct {
	Raw  Bytes
	Node *LinkedNode
}

/**
 * Output
 */
type Output struct {
	Exp  *lexer.Expression
	Node *LinkedNode
}

/**
 * Assignment
 */
type Assignment struct {
	Scope InScope
	Name  lexer.IToken
	Value *lexer.Expression
}

/**
 * Include
 */
type Include struct {
	Depth uint
	Node  *LinkedNode
	Tpl   *Template
}

/**
 * Extend
 */
type Extend struct {
	Tpl     *Template
	TplPath *TplPath
}

/**
 * Slot
 */
type Slot struct {
	Name  Bytes
	Node  *LinkedNode
	Codes []ICode
}

/**
 * Template
 */
type Template struct {
	TplPath *TplPath
	Codes   []ICode
}
