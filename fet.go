package fet

import (
	"fmt"

	"github.com/fefit/fet/lexer"
	"github.com/fefit/fet/utils"
)

type InScope = uint
type CodeType = uint
type PathType = uint
type Bytes = []byte

const (
	LocalScope InScope = iota
	ParentScope
	GlobalScope
)

const (
	BlockCode CodeType = iota
	BlockFeatureCode
	BlockEndCode
	UnkownCode
	DetectCode
	CommentCode
	RawHtmlCode
	OutputCode
	AssignCode
)

const (
	AbsPath  PathType = iota // absolute path
	RelaPath                 // relative path
	BasePath                 // path base on
)

func createBytesMatcher(bytes *Bytes, spaceIndex int) *BytesMatcher {
	return &BytesMatcher{
		Index:      0,
		SpaceIndex: spaceIndex,
		Raw:        bytes,
	}
}

type BytesMatcher struct {
	IsEnd      bool
	Index      int
	SpaceIndex int
	Raw        *Bytes
}

func (matcher *BytesMatcher) Match(bt byte) (bool, error) {
	index := matcher.Index
	raw := *matcher.Raw
	curByte := raw[index]
	if bt == curByte {
		if index == len(raw)-1 {
			matcher.IsEnd = true
			return true, nil
		}
		matcher.Index++
		return false, nil
	}
	// allow space
	if index == matcher.SpaceIndex && utils.IsSpaceByte(bt) {
		return false, nil
	}
	return false, fmt.Errorf("the byte '%s' does not match byte '%s'", string(bt), string(curByte))
}

type TplPath struct {
	Type    PathType
	Path    []Bytes
	AbsPath []Bytes
}

type Config struct {
	LeftDelimiter  Bytes
	RightDelimiter Bytes
	CommentSymbol  byte
	RegisterBlocks []*RegisterBlock
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
	Type() CodeType
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

func (unkown *Unkown) Type() CodeType {
	return UnkownCode
}

/**
 *
 */
type Detect struct {
	Parsed      Bytes
	ProxyHandle *func(bt byte, parser *Parser) (ICode, error)
	Code        ICode
	EndMatched  int
}

func (detect *Detect) HandleMaybeBlockOrBlockFeature(bt byte, parser *Parser) (ICode, error) {
	// maybe block names
	if lexer.IsAlphaByte(bt) {
		detect.Parsed = append(detect.Parsed, bt)
		return nil, nil
	}
	// maybe block
	endDelim := parser.Config.RightDelimiter
	endMatched := detect.EndMatched
	isEnd := false
	if bt == endDelim[detect.EndMatched] {
		detect.EndMatched++
		isEnd = detect.EndMatched == len(endDelim)
	} else {
		detectedCode := detect.Code
		if detectedCode != nil {
			// has detected the code type
			if endMatched > 0 {
				for _, endBt := range endDelim[:endMatched] {
					if curCode, err := detectedCode.Add(endBt, parser); err != nil {
						return curCode, err
					}
				}
				// reset end matched
				detect.EndMatched = 0
			}
			return detectedCode.Add(bt, parser)
		} else {
			parsed := detect.Parsed
			// the first byte
			var firstByte byte
			if endMatched > 0 {
				firstByte = endDelim[0]
			} else {
				firstByte = bt
			}
			// step1: check if block name
			if registerBlock := parser.MatchBlock(parsed); registerBlock != nil {
				// block
				maybeBlock := &Block{
					Meta: registerBlock,
				}
				// add next byte after block name
				if curCode, err := maybeBlock.Add(firstByte, parser); err == nil {
					// match the block
					if endMatched > 0 {

					}
				} else {
					// take it as output

				}
			} else if utils.IsSpaceByte(firstByte) {
				// maybe block feature
				if block, isBlock := parser.CurCode.(*Block); isBlock && len(block.Meta.Features) > 0 {

				}
			}
		}
	}

	return nil, nil
}

func (detect *Detect) HandleOutput(bt byte, parser *Parser) (ICode, error) {
	return nil, nil
}

func (detect *Detect) HandleMaybeOutputOrAssign(bt byte, parser *Parser) (ICode, error) {
	return nil, nil
}

func (detect *Detect) Add(bt byte, parser *Parser) (ICode, error) {
	handle := detect.ProxyHandle
	if handle == nil {
		// ignore whitespace
		if utils.IsSpaceByte(bt) {
			return nil, nil
		}
		// comment code
		if bt == parser.Config.CommentSymbol {
			return &Comment{}, nil
		}
		// block end
		if bt == '/' {
			// end block
			if block, isBlock := parser.CurCode.(*Block); isBlock {
				*handle = block.AddEnd
				return nil, nil
			}
			// wrong end block
			return nil, fmt.Errorf("wrong block end /")
		}
		// detect which handle should use
		if utils.IsEnLetterByte(bt) {
			// maybe block/block feature
			*handle = detect.HandleMaybeBlockOrBlockFeature
		} else {
			// parse into expression
			if bt == '$' {
				// maybe output or assignment
				*handle = detect.HandleMaybeOutputOrAssign
			} else {
				// must be output
				*handle = detect.HandleOutput
			}
		}
	}
	return (*handle)(bt, parser)
}

func (_ *Detect) Type() CodeType {
	return DetectCode
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

type RegisterBlockFeature struct {
	Once       bool  // only allowed appear once
	Last       bool  // only allowed appear at last feature
	SpaceIndex int   // if allow whitespace in keywords e.g. else if => elseif
	Name       Bytes // feature block name
	Parser     IParser
}
type BlockFeature struct {
	Meta *RegisterBlockFeature
	Name Bytes
}

/**
 *
 */
type RegisterBlock struct {
	Name     Bytes
	Features []RegisterBlockFeature
	Parser   IParser
}

type Block struct {
	NameEnded bool
	Meta      *RegisterBlock
	Node      *LinkedNode
	End       *BytesMatcher
}

func (block *Block) Add(bt byte, parser *Parser) (ICode, error) {
	return nil, nil

}

func (_ *Block) Type() CodeType {
	return BlockCode
}

func (block *Block) AddEnd(bt byte, parser *Parser) (ICode, error) {
	isEnd, err := block.End.Match(bt)
	if err != nil {
		return block, fmt.Errorf("wrong end block ''")
	}
	// the block name is ended or the block is closed
	if isEnd {
		if block.NameEnded {
			// the block is closed

		} else {
			// the block name end
			block.NameEnded = true
			// next should be right delimiter
			// and allow the beginning bytes are spaces
			block.End = createBytesMatcher(&parser.Config.RightDelimiter, 0)
		}
	}
	return nil, nil
}

type BlockEnd struct {
	Name  *Bytes
	Index uint
}

func (blockEnd *BlockEnd) Add(bt byte, parser *Parser) (ICode, error) {
	return nil, nil
}

func (blockEnd *BlockEnd) Type() CodeType {
	return BlockEndCode
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

func (_ *RawHtml) Type() CodeType {
	return RawHtmlCode
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

func (_ *Comment) Type() CodeType {
	return CommentCode
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

func (parser *Parser) MatchBlock(name []byte) *RegisterBlock {
	for _, block := range parser.Config.RegisterBlocks {
		if utils.IsSameBytes(block.Name, name) {
			return block
		}
	}
	return nil
}

func (parser *Parser) Parse(bt byte) error {

	return nil
}

type Fet struct {
}
