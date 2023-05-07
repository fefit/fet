/**
 * The lexer for parsing fet template
 */
package lexer

import (
	"fmt"
)

type TokenType uint8
type NumberBase uint8

const (
	OpType TokenType = iota
	IdentType
	SpaceType
	NumType
	StrType
	ArrLitType
	FuncCallType
	PipeFuncType
	ObjPropType
	ArrFieldType
)

const (
	BYTE_UNDERSCORE = '_'
	BYTE_SPACE      = ' '
)

const (
	HexBase     NumberBase = 16
	OctalBase   NumberBase = 8
	BinaryBase  NumberBase = 2
	DecimalBase NumberBase = 10
)

/**
 * space
 */
func IsSpaceByte(bt byte) bool {
	return bt == BYTE_SPACE || bt == '\t'
}

/**
 *
 */
func IsAlphaByte(bt byte) bool {
	return (bt >= 'a' && bt <= 'z') || (bt >= 'A' && bt <= 'Z')
}

/**
 * number byte functions
 */
func IsDecimalByte(bt byte) bool {
	return bt >= '0' && bt <= '9'
}

func IsOctalByte(bt byte) bool {
	return bt >= '0' && bt < '8'
}

func IsBinaryByte(bt byte) bool {
	return bt == '0' || bt == '1'
}

func IsHexByte(bt byte) bool {
	return (bt >= 'a' && bt <= 'f') || (bt >= 'A' && bt <= 'F') || IsDecimalByte(bt)
}

func IsBaseNumberByte(bt byte, base NumberBase) bool {
	switch base {
	case DecimalBase:
		return IsDecimalByte(bt)
	case HexBase:
		return IsHexByte(bt)
	case OctalBase:
		return IsOctalByte(bt)
	case BinaryBase:
		return IsBinaryByte(bt)
	}
	return false
}

func AddSpaceOrOperatorByte(bt byte) (IToken, error) {
	if IsSpaceByte(bt) {
		return &SpaceToken{
			Raw: []byte{bt},
		}, nil
	}
	// array literal
	if bt == '[' {
		return &bracketOperator, nil
	}
	// paren
	if bt == '(' {
		return &parenOperator, nil
	}
	// operators
	for _, op := range Operators {
		if bt == op.Raw[0] {
			return &op, nil
		}
	}
	// also not an operator
	return nil, fmt.Errorf("only allowed operator or space token")
}

func AddUnkownTokenByte(bt byte, prevToken IToken) (IToken, error) {
	// space or operator
	if token, err := AddSpaceOrOperatorByte(bt); err == nil {
		return token, nil
	}
	// double quote
	if bt == '"' {
		return &DoubleQuoteStringToken{}, nil
	}
	// single quote
	if bt == '\'' {
		return &SingleQuoteStringToken{}, nil
	}
	// try number token
	num := NewNumberToken()
	if _, err := num.Add(bt, prevToken); err == nil {
		return num, nil
	}
	// identifier token
	ident := &IdentifierToken{}
	if _, err := ident.Add(bt, prevToken); err == nil {
		return ident, nil
	}
	// not allowed
	return nil, fmt.Errorf("syntax error: '%s' is not a correct token.", string(bt))
}

func IsParenOpToken(token IToken) bool {
	if op, isOp := token.(*OperatorToken); isOp && op == &parenOperator {
		return true
	}
	return false
}

func IsPipeOpToken(token IToken) bool {
	if op, isOp := token.(*OperatorToken); isOp && op.Raw[0] == '|' && len(op.Raw) == 1 {
		return true
	}
	return false
}

/**
 * -------------------------
 * Interfaces
 */
/**
 * Token types which can be the key of an array
 */
type IToken interface {
	Add(bt byte, prevToken IToken) (IToken, error)
	Type() TokenType
	End() error
}

var Operators = []OperatorToken{
	{
		Raw:      []byte("+"), // Addition
		Priority: 11,
		NextMaybe: &OperatorToken{
			Raw:      []byte("+"), // Unary plus
			Unary:    true,
			Priority: 14,
			NextMaybe: &OperatorToken{
				Raw:      []byte("++"), // Postfix Increment
				Priority: 14,
				Unary:    true,
				NextMaybe: &OperatorToken{
					Raw:         []byte("++"), // Prefix Increment
					Priority:    15,
					Unary:       true,
					RightToLeft: true,
				},
			},
		},
	},
	{
		Raw:      []byte("-"), // Subtraction
		Priority: 11,
		NextMaybe: &OperatorToken{
			Raw:      []byte("-"), // Unary negation
			Priority: 14,
			Unary:    true,
			NextMaybe: &OperatorToken{
				Raw:      []byte("--"), // Postfix Decrement
				Priority: 14,
				Unary:    true,
				NextMaybe: &OperatorToken{
					Raw:         []byte("--"), // Prefix Decrement
					Priority:    15,
					Unary:       true,
					RightToLeft: true,
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
}

var parenOperator = OperatorToken{
	Raw:      []byte("("),
	Priority: 18,
}

var bracketOperator = OperatorToken{
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
var spaceToken = SpaceToken{}

/**
 * Space Token
 */
type SpaceToken struct {
	Raw []byte
}

func (sp *SpaceToken) Add(bt byte, prevToken IToken) (IToken, error) {
	if IsSpaceByte(bt) {
		sp.Raw = append(sp.Raw, bt)
		return nil, nil
	}
	return AddUnkownTokenByte(bt, prevToken)
}

func (sp *SpaceToken) Type() TokenType {
	return SpaceType
}

func (sp *SpaceToken) End() error {
	return nil
}

/**
 * Number token
 */
type Integer struct {
	PrevByte *byte
	Raw      []byte
}

func (it *Integer) CheckPrevByte() (IToken, error) {
	if it.PrevByte == nil {
		return nil, fmt.Errorf("syntax error: empty")
	}
	if *it.PrevByte == BYTE_UNDERSCORE {
		return nil, fmt.Errorf("syntax error: unexpect _")
	}
	return nil, nil
}

type NumberToken struct {
	BeginWithZero bool
	Base          NumberBase
	Integer       Integer
	Decimal       *Integer
	Exponent      *Integer
}

func NewNumberToken() *NumberToken {
	return &NumberToken{
		Base: DecimalBase,
	}
}

func (num *NumberToken) Add(bt byte, prevToken IToken) (IToken, error) {
	var curInteger *Integer
	// check the base
	if num.Base != DecimalBase {
		// if the base is not 10
		curInteger = &num.Integer
	} else {
		// keep the order
		if num.Exponent != nil {
			// first, check the exponent part
			// the first byte maybe the symbol '-' or '+'
			if len(num.Exponent.Raw) == 0 && (bt == '-' || bt == '+') {
				num.Exponent.Raw = append(num.Exponent.Raw, bt)
				// exponent with prefix symbol
				return nil, nil
			}
			// check the exponent
			curInteger = num.Exponent
		} else if num.Decimal != nil {
			// second, check the decimal part
			// check if is exponent
			if bt == 'e' || bt == 'E' {
				num.Exponent = &Integer{}
				return num.Decimal.CheckPrevByte()
			}
			curInteger = num.Decimal
		} else {
			// then check the integer part
			intNum := len(num.Integer.Raw)
			if intNum == 0 {
				if bt == '0' {
					// begin with zero
					num.BeginWithZero = true
					// also set prev byte
					num.Integer.PrevByte = &bt
				} else if bt >= '1' && bt <= '9' {
					// normal number
					num.Integer.PrevByte = &bt
				} else {
					// wrong number token
					return nil, fmt.Errorf("wrong number token")
				}
				// add byte to integer
				num.Integer.Raw = append(num.Integer.Raw, bt)
				// byte is ok
				return nil, nil
			} else {
				if bt == '.' {
					// float
					num.Decimal = &Integer{}
					// check prev byte
					return num.Integer.CheckPrevByte()
				} else if bt == 'e' || bt == 'E' {
					// exponent
					num.Exponent = &Integer{}
					// check prev byte
					return num.Integer.CheckPrevByte()
				} else {
					// begin with zero
					if intNum == 1 && num.BeginWithZero {
						if bt == 'x' || bt == 'X' {
							// hex number
							num.Base = HexBase
						} else if bt == 'o' || bt == 'O' {
							// octal number
							num.Base = OctalBase
						} else if bt == 'b' || bt == 'B' {
							// binaray number
							num.Base = BinaryBase
						} else {
							// take it as octal number
							if IsOctalByte(bt) {
								num.Base = OctalBase
							} else if bt == BYTE_UNDERSCORE {
								num.Base = OctalBase
								num.Integer.PrevByte = &bt
							} else {
								// wrong octal number
								return nil, fmt.Errorf("wrong octal number")
							}
						}
						// add byte to integer
						num.Integer.Raw = append(num.Integer.Raw, bt)
						// byte is ok
						return nil, nil
					} else {
						// still in integer
						curInteger = &num.Integer
					}
				}
			}
		}
	}
	// check current integer
	if bt == BYTE_UNDERSCORE {
		// not allowed '_' appear at the beginning or repeated _
		if curInteger.PrevByte == nil || *curInteger.PrevByte == BYTE_UNDERSCORE {
			return nil, fmt.Errorf("syntax error:")
		}
	} else {
		// determine whether the current byte belongs to the base
		if !IsBaseNumberByte(bt, num.Base) {
			if _, err := curInteger.CheckPrevByte(); err != nil {
				return nil, err
			} else {
				// the number token is end
				return AddSpaceOrOperatorByte(bt)
			}
		}
	}
	// reset the prev byte
	curInteger.PrevByte = &bt
	curInteger.Raw = append(curInteger.Raw, bt)
	// return ok
	return nil, nil
}

func (num *NumberToken) Type() TokenType {
	return NumType
}

func (num *NumberToken) End() error {
	var curInteger *Integer
	if num.Exponent != nil {
		curInteger = num.Exponent
	} else if num.Decimal != nil {
		curInteger = num.Decimal
	} else {
		curInteger = &num.Integer
	}
	_, err := curInteger.CheckPrevByte()
	return err
}

/**
 * Identifier token, e.g $a abc a123
 */
type IdentifierToken struct {
	IsVariable bool
	Raw        []byte
}

func (id *IdentifierToken) Add(bt byte, prevToken IToken) (IToken, error) {
	if len(id.Raw) == 0 {
		if bt == '$' {
			id.IsVariable = true
		} else {
			if bt == BYTE_UNDERSCORE || IsAlphaByte(bt) {
				// allowed identifier bytes
			} else {
				// not an identifier
				return nil, fmt.Errorf("not an identifier")
			}
		}
	} else {
		if IsAlphaByte(bt) || IsDecimalByte(bt) || bt == BYTE_UNDERSCORE {
			// ok
		} else {
			// check if is variable and only a $
			if len(id.Raw) == 1 && id.IsVariable {
				return nil, fmt.Errorf("wrong variable $")
			}
			// next only allow space or operator
			return AddSpaceOrOperatorByte(bt)
		}
	}
	id.Raw = append(id.Raw, bt)
	return nil, nil
}

func (id *IdentifierToken) Type() TokenType {
	return IdentType
}

func (id *IdentifierToken) End() error {
	return nil
}

/**
 * string token
 * SingleQuote: 'abc'
 * DoubleQuote: "abc${abc}"
 * Tempalte: `"abc"${abc}`
 */
// Single Quote String
type SingleQuoteStringToken struct {
	InTranslate bool
	IsEnd       bool
	Raw         []byte
}

func (ss *SingleQuoteStringToken) Add(bt byte, prevToken IToken) (IToken, error) {
	// first check if the string has been end
	if ss.IsEnd {
		return AddSpaceOrOperatorByte(bt)
	}
	// check if is in translate
	if ss.InTranslate {
		ss.InTranslate = false
	} else {
		// translate
		if bt == '\\' {
			ss.InTranslate = true
		} else if bt == '\'' {
			// end '
			ss.IsEnd = true
			return nil, nil
		}
	}
	ss.Raw = append(ss.Raw, bt)
	return nil, nil
}

func (ss *SingleQuoteStringToken) Type() TokenType {
	return StrType
}

func (ss *SingleQuoteStringToken) End() error {
	if !ss.IsEnd {
		return fmt.Errorf("the single quote string is not closed")
	}
	return nil
}

// Double Quote String
type RepExp struct {
	Range []uint
	Exp   Expression
}
type DoubleQuoteStringToken struct {
	InTranslate bool
	IsEnd       bool
	InExp       bool
	Raw         []byte
	Exps        []RepExp
}

func (ds *DoubleQuoteStringToken) Add(bt byte, prevToken IToken) (IToken, error) {
	// check if is end
	if ds.IsEnd {
		return AddSpaceOrOperatorByte(bt)
	}
	// check if is in translate
	if ds.InTranslate {
		ds.InTranslate = false
	} else {
		if bt == '\\' {
			// translate
			ds.InTranslate = true
		} else if bt == '`' {
			// expression
			ds.InExp = true
		} else if bt == '"' {
			// end '
			ds.IsEnd = true
			return nil, nil
		}
	}
	ds.Raw = append(ds.Raw, bt)
	return nil, nil
}

func (ds *DoubleQuoteStringToken) Type() TokenType {
	return StrType
}

func (ds *DoubleQuoteStringToken) End() error {
	if !ds.IsEnd {
		return fmt.Errorf("the double quote string is not closed")
	}
	return nil
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

func (op *OperatorToken) Add(bt byte, prevToken IToken) (IToken, error) {
	var curByteLen = len(op.Raw)
	var next *OperatorToken = op
	for {
		next = next.NextMaybe
		if next == nil {
			break
		}
		if len(next.Raw) > curByteLen && next.Raw[curByteLen] == bt {
			// change op to the next
			return next, fmt.Errorf("should use the max matched operator")
		}
	}
	return AddUnkownTokenByte(bt, prevToken)
}

func (op *OperatorToken) Type() TokenType {
	return OpType
}

func (op *OperatorToken) End() error {
	return nil
}

/**
 * Array Literal
 * e.g => [ "a" => 1, "c"]
 */
type ArrayLiteral struct{}

func (arr *ArrayLiteral) Add(bt byte, prevToken IToken) (IToken, error) {
	return nil, nil
}

func (arr *ArrayLiteral) Type() TokenType {
	return ArrLitType
}

func (arr *ArrayLiteral) End() error {
	return nil
}

/**
 * Function Call
 * e.g => abc(), abc(1+2, "def")
 */
type FunctionCall struct {
	Name IToken
	Args []Expression
}

func (fn *FunctionCall) Add(bt byte, prevToken IToken) (IToken, error) {
	return nil, nil
}

func (fn *FunctionCall) Type() TokenType {
	return FuncCallType
}

func (fn *FunctionCall) End() error {
	return nil
}

/**
*
 */
type PipeFunction struct {
	Name IdentifierToken
	Args []IToken
}

func (pipe *PipeFunction) Add(bt byte, prevToken IToken) (IToken, error) {
	return nil, nil
}

func (pipe *PipeFunction) Type() TokenType {
	return PipeFuncType
}

func (pipe *PipeFunction) End() error {
	return nil
}

/*
* Object Property
 */
type ObjectProperty struct {
	Object   IToken
	Property IdentifierToken
}

func (obj *ObjectProperty) Add(bt byte, prevToken IToken) (IToken, error) {
	return nil, nil
}

func (obj *ObjectProperty) Type() TokenType {
	return ObjPropType
}

func (obj *ObjectProperty) End() error {
	return nil
}

/**
*
 */

type ArrayField struct {
	Array        IToken
	Field        *IdentifierToken
	DynamicField *Expression
}

func (obj *ArrayField) Add(bt byte, prevToken IToken) (IToken, error) {
	return nil, nil
}

func (obj *ArrayField) Type() TokenType {
	return ArrFieldType
}

func (obj *ArrayField) End() error {
	return nil
}

type Ast struct {
	Op    *OperatorToken
	Left  IToken
	Right IToken
}

type Expression struct {
	PrevToken IToken
	CurToken  IToken
	OpStack   []*OperatorToken
	Output    []IToken
}

func New() Expression {
	return Expression{
		CurToken:  &spaceToken,
		PrevToken: &spaceToken,
	}
}

func (exp *Expression) getPrevToken() IToken {
	if exp.CurToken.Type() == SpaceType {
		return exp.PrevToken
	}
	return exp.CurToken
}

func (exp *Expression) Add(bt byte) (IToken, error) {
	var curToken = exp.CurToken
	if nextToken, err := curToken.Add(bt, exp.PrevToken); err == nil {
		// change current token to next
		// if next token is not nil
		if nextToken != nil {
			fmt.Printf("%#v", curToken)
			fmt.Println()
			// special
			if op, isOp := nextToken.(*OperatorToken); isOp {
				if op == &parenOperator {
					// '('
					prevType := exp.CurToken.Type()
					if prevType == IdentType || prevType == FuncCallType || prevType == ObjPropType || prevType == ArrFieldType {
						// check if function call
						nextToken = &FunctionCall{}
					} else {
						// normal group operator '('
					}
				} else if op == &bracketOperator {
					// '[', if before is '(', take it as a array literal type
					if IsParenOpToken(exp.getPrevToken()) {
						nextToken = &ArrayLiteral{}
					} else {
						// '[' dynamic array field
						nextToken = &ArrayField{}
					}
				}
			} else if ident, isIdent := nextToken.(*IdentifierToken); isIdent && !ident.IsVariable && IsPipeOpToken(exp.getPrevToken()) {
				// pipe function, a|count
				nextToken = &PipeFunction{}
			}
			// if not space token,
			if exp.CurToken.Type() != SpaceType {
				exp.PrevToken = exp.CurToken
			}
			// set current token
			exp.CurToken = nextToken
		}
		return exp.CurToken, nil
	} else if nextToken != nil {
		// should replace the previous token
		exp.CurToken = nextToken
		return nextToken, nil
	} else {
		return nil, err
	}
}

func (exp *Expression) Eof() error {
	return nil
}

func (exp *Expression) Parse(str string) (*Ast, error) {
	for i := 0; i < len(str); i++ {
		if _, err := exp.Add(str[i]); err != nil {
			return nil, err
		}
	}
	if err := exp.Eof(); err != nil {
		return nil, err
	}
	return &Ast{}, nil
}
