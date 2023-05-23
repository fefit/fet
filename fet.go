package fet

import (
	"unicode"

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
	CommentSymbol  byte
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
	Add(bt byte, parser *Parser) (ICode, error)
}

type LinkedNode struct {
	Parent     ICode // parent node
	Prev       ICode // prev node
	Next       ICode // next node
	FirstChild ICode // first child
}

/**
 * Initital State Code
 */
type Unkown struct {
	Matched int
}

func (unkown *Unkown) Add(bt byte, parser *Parser) (ICode, error) {
	leftDelim := parser.Config.LeftDelimiter
	index := unkown.Matched
	if bt == leftDelim[index] {
		if index == len(leftDelim)-1 {
			// matched all the left delimiter
			return parser.Detect, nil
		}
		// increase the matched count
		unkown.Matched++
		return nil, nil
	}
	if index > 0 {
		// non full-matched delimiter
		raw := &RawHtml{
			Raw: append([]byte{}, leftDelim[:index]...),
		}
		return raw.Add(bt, parser)
	}
	return &RawHtml{
		Raw: []byte{bt},
	}, nil
}

type Detect struct {
	Raw Bytes
}

func (detect *Detect) Add(bt byte, parser *Parser) (ICode, error) {
	raw := detect.Raw
	total := len(raw)
	if total == 0 {
		if bt == parser.Config.CommentSymbol {
			return &Comment{}, nil
		}
		if !unicode.IsSpace(rune(bt)) {
			detect.Raw = []byte{bt}
		}
	} else {

	}
	return nil, nil
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

func (rawHtml *RawHtml) Add(bt byte, parser *Parser) (ICode, error) {
	return nil, nil

}

/**
 *
 */
type Comment struct {
	Raw Bytes
}

func (comment *Comment) Add(bt byte, parser *Parser) (ICode, error) {
	return nil, nil

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

/**
 *
 */
type Parser struct {
	Config  *Config // Config
	Unkown  *Unkown // Unkown code
	Detect  *Detect // Detect block/assign/output
	CurCode ICode   // Current code
}

func (parser *Parser) Parse(bt byte) error {

	return nil
}
