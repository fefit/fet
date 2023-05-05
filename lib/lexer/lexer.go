/**
 * The lexer for parsing fet template
 */
package lexer

import "fmt"

type TokenType uint8

const (
	ProxyType TokenType = iota
	OpType
	IdentType
	SpaceType
	NumType
	StrType
	ArrType
	FuncType
)

/**
 * -------------------------
 * Interfaces
 */
/**
 * Token types which can be the key of an array
 */
type IToken interface {
	Add(bt byte) (IToken, error)
	Type() TokenType
}

var Operators = []OperatorToken{
	{
		Raw:      []byte("+"), // Unary plus
		Priority: 14,
		Unary:    true,
		NextMaybe: &OperatorToken{
			Raw:      []byte("+"), // Addition
			Priority: 11,
			NextMaybe: &OperatorToken{
				Raw:         []byte("++"), // Postfix Increment
				Priority:    15,
				Unary:       true,
				RightToLeft: true,
				NextMaybe: &OperatorToken{
					Raw:      []byte("++"), // Prefix Increment
					Priority: 14,
					Unary:    true,
				},
			},
		},
	},
	{
		Raw:      []byte("-"), // Unary negation
		Priority: 14,
		Unary:    true,
		NextMaybe: &OperatorToken{
			Raw:      []byte("-"), // Subtraction
			Priority: 11,
			NextMaybe: &OperatorToken{
				Raw:         []byte("--"), // Postfix Decrement
				Priority:    15,
				Unary:       true,
				RightToLeft: true,
				NextMaybe: &OperatorToken{
					Raw:      []byte("--"), // Prefix Decrement
					Priority: 14,
					Unary:    true,
					NextMaybe: &OperatorToken{
						Raw:      []byte("->"), // Object Member Access
						Priority: 17,
					},
				},
			},
		},
	}, {
		Raw:      []byte("*"), // Multiplication
		Priority: 12,
		NextMaybe: &OperatorToken{
			Raw:         []byte("**"), // Exponentiation
			Priority:    13,
			RightToLeft: true,
		},
	},
	{
		Raw:      []byte("/"), // Division
		Priority: 12,
	}, {
		Raw:      []byte("%"), // Remainder
		Priority: 12,
	},
	{
		Raw:        []byte("!"), // Logic Not
		Priority:   14,
		Unary:      true,
		RepeatAble: true,
		NextMaybe: &OperatorToken{
			Raw:      []byte("!="), // Inequality
			Priority: 8,
		},
	},
	{
		Raw:      []byte(">"), // Greater Than
		Priority: 9,
		NextMaybe: &OperatorToken{
			Raw:      []byte(">="), // Greate Than Or Equal
			Priority: 9,
			NextMaybe: &OperatorToken{
				Raw:      []byte(">>"), // Bitwise Right Shift
				Priority: 10,
			},
		},
	},
	{
		Raw:      []byte("<"), // Less Than
		Priority: 9,
		NextMaybe: &OperatorToken{
			Raw:      []byte("<="), // Less Than Or Equal
			Priority: 9,
			NextMaybe: &OperatorToken{
				Raw:      []byte("<<"), // Bitwise Left Shift
				Priority: 10,
			},
		},
	},
	{
		Raw:      []byte("."), // Member Access
		Priority: 17,
	}, {
		Raw:      []byte("&"), // Bitwise And
		Priority: 7,
		NextMaybe: &OperatorToken{
			Raw:      []byte("&&"), // Logic And
			Priority: 4,
		},
	},
	{
		Raw:      []byte("~"), // Bitwise Not
		Priority: 14,
	},
	{
		Raw:      []byte("^"), // Bitwise XOR
		Priority: 6,
	},
	{
		Raw:      []byte("|"), // Bitwise Or
		Priority: 5,
		NextMaybe: &OperatorToken{
			Raw:      []byte("||"), // Logic Or
			Priority: 3,
		},
	},
	{
		Raw:      []byte(","), // Comma
		Priority: 1,
	},
}

var ParenOperator = OperatorToken{
	Raw:      []byte("("),
	Priority: 18,
}

var BracketOperator = OperatorToken{
	Raw:      []byte("["),
	Priority: 17,
}

/**
 * ------------------------------
 * token types
 */
/**
 * Proxy token
 */
var proxyToken = ProxyToken{}

type ProxyToken struct {
}

func (proxy *ProxyToken) Add(bt byte) (IToken, error) {
	// double quote
	if bt == '"' {
		return &DoubleQuoteStringToken{}, nil
	}
	// single quote
	if bt == '\'' {
		return &SingleQuoteStringToken{}, nil
	}
	// number token
	if bt >= '0' && bt <= '9' {
		return &NumberToken{
			Raw: []byte{bt},
		}, nil
	}
	// space
	if bt == ' ' || bt == '\t' {
		return &SpaceToken{
			Raw: []byte{bt},
		}, nil
	}
	// array literal
	if bt == '[' {
		return &ArrayLiteral{}, nil
	}
	// paren
	if bt == '(' {
		return &ParenOperator, nil
	}
	// operators
	for _, op := range Operators {
		if op.Unary {
			if bt == op.Raw[0] {
				return &op, nil
			}
		}
	}
	// not allowed
	return nil, fmt.Errorf("syntax error: '%s' is not a correct token.", string(bt))
}

func (proxy *ProxyToken) Type() TokenType {
	return ProxyType
}

/**
 *
 */
type SpaceToken struct {
	Raw []byte
}

func (sp *SpaceToken) Add(bt byte) (IToken, error) {
	return nil, nil
}
func (sp *SpaceToken) Type() TokenType {
	return SpaceType
}

/**
 * Number token
 */
type NumberToken struct {
	Raw []byte
}

func (num *NumberToken) Add(bt byte) (IToken, error) {
	return nil, nil
}

func (num *NumberToken) Type() TokenType {
	return NumType
}

/**
 * Identifier token, e.g $a abc a123
 */
type IdentifierToken struct {
	IsKeyword bool
	Value     []byte
	Raw       []byte
}

func (id *IdentifierToken) Add(bt byte) (IToken, error) {
	return nil, nil
}

func (id *IdentifierToken) Type() TokenType {
	return IdentType
}

/**
 * string token
 * SingleQuote: 'abc'
 * DoubleQuote: "abc${abc}"
 * Tempalte: `"abc"${abc}`
 */
// Single Quote String
type SingleQuoteStringToken struct {
	Raw []byte
}

func (ss *SingleQuoteStringToken) Add(bt byte) (IToken, error) {
	return nil, nil
}

func (ss *SingleQuoteStringToken) Type() TokenType {
	return StrType
}

// Double Quote String
type DoubleQuoteStringToken struct {
}

func (ds *DoubleQuoteStringToken) Add(bt byte) (IToken, error) {
	return nil, nil
}

func (ds *DoubleQuoteStringToken) Type() TokenType {
	return StrType
}

// Template String
type TemplateStringToken struct {
}

func (ts *TemplateStringToken) Add(bt byte) (IToken, error) {
	return nil, nil
}

func (ts *TemplateStringToken) Type() TokenType {
	return StrType
}

// Raw String
type RawStringToken struct{}

func (rs *RawStringToken) Add(bt byte) (IToken, error) {
	return nil, nil
}

func (rs *RawStringToken) Type() TokenType {
	return StrType
}

/**
 * operator token
 */
type OperatorToken struct {
	Unary       bool
	RightToLeft bool
	RepeatAble  bool
	Priority    uint8
	Raw         []byte
	NextMaybe   *OperatorToken
}

func (op *OperatorToken) Add(bt byte) (IToken, error) {
	return nil, nil
}

func (op *OperatorToken) Type() TokenType {
	return OpType
}

/**
 * Array Literal
 * e.g => [ "a" => 1, "c"]
 */
type ArrayLiteral struct{}

func (arr *ArrayLiteral) Add(bt byte) (IToken, error) {
	return nil, nil
}

func (arr *ArrayLiteral) Type() TokenType {
	return ArrType
}

/**
 * Function Call
 * e.g => abc(), abc(1+2, "def")
 */
type FunctionCall struct {
}

func (fn *FunctionCall) Add(bt byte) (IToken, error) {
	return nil, nil
}

func (fn *FunctionCall) Type() TokenType {
	return FuncType
}

type Ast struct {
	Op    *OperatorToken
	Left  *IToken
	Right *IToken
}

type Expression struct {
	PrevToken IToken
	CurToken  IToken
	OpStack   []*OperatorToken
	Output    []IToken
	Ast       Ast
}

func New() Expression {
	return Expression{
		CurToken: &proxyToken,
	}
}

func (exp *Expression) PrevIsOp() bool {
	_, isOp := exp.PrevToken.(*OperatorToken)
	return isOp
}

func (exp *Expression) Add(bt byte) (IToken, error) {
	var curToken = exp.CurToken
	if nextToken, err := curToken.Add(bt); err == nil {
		// change current token to next
		// if next token is not nil
		if nextToken != nil {
			// judge operator
			if op, isOp := curToken.(*OperatorToken); isOp {

			}
			exp.CurToken = nextToken
		}
		return exp.CurToken, nil
	} else {
		return nil, err
	}
}
