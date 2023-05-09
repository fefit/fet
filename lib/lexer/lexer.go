/**
 * The lexer for parsing fet template
 */
package lexer

import (
	"bytes"
	"fmt"

	"github.com/davecgh/go-spew/spew"
)

type TokenType uint8
type NumberBase uint8
type ArrayInState uint8

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
	ExpType
	AstType
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

const (
	MaybeArrayKey ArrayInState = iota
	WaitGreatThan
	InArrayValue
)

/**
 * space
 */
func IsSpaceByte(bt byte) bool {
	return bt == BYTE_SPACE || bt == '\t'
}

func IsSpaceToken(token IToken) bool {
	return token.Type() == SpaceType
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
		return &OperatorToken{
			Op: &bracketOperator,
		}, nil
	}
	// paren
	if bt == '(' {
		return &OperatorToken{
			Op: &parenOperator,
		}, nil
	}
	// member
	if bt == '.' {
		return &OperatorToken{
			Op: &memberOperator,
		}, nil
	}
	// paren end
	if bt == ')' {
		return &OperatorToken{
			Op: &parenEndOperator,
		}, nil
	}
	// operators
	for _, op := range operatorList {
		if bt == op.Raw[0] {
			return &OperatorToken{
				Op: &op,
			}, nil
		}
	}
	// also not an operator
	return nil, fmt.Errorf("only allowed operator or space token: %s", string(bt))
}

func AddUnkownTokenByte(bt byte, exp *Expression) (IToken, error) {
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
	if _, err := num.Add(bt, exp); err == nil {
		return num, nil
	}
	// identifier token
	ident := &IdentifierToken{}
	if _, err := ident.Add(bt, exp); err == nil {
		return ident, nil
	}
	// not allowed
	return nil, fmt.Errorf("syntax error: '%s' is not a correct token.", string(bt))
}

func IsOpToken(token IToken, op *Operator) bool {
	if opToken, isOp := token.(*OperatorToken); isOp && opToken.Op == op {
		return true
	}
	return false
}

func IsPipeOpToken(token IToken) bool {
	if opToken, isOp := token.(*OperatorToken); isOp && opToken.Op.Raw[0] == '|' && len(opToken.Op.Raw) == 1 {
		return true
	}
	return false
}

func IsSameBytes(a []byte, b []byte) bool {
	total := len(a)
	if total == len(b) {
		for i := 0; i < total; i++ {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}
	return false
}

/**
 * Operator
 */
type Operator struct {
	Unary       bool
	RightToLeft bool
	Priority    uint8
	Raw         []byte
	NextMaybe   *Operator
}

func (op *Operator) FixIfUnary(prevToken IToken) (*Operator, error) {
	nextOp := op.NextMaybe
	if nextOp != nil && IsSameBytes(op.Raw, nextOp.Raw) {
		prevType := prevToken.Type()
		if prevType == SpaceType {
			// at the beginning of the expression
			return nextOp, nil
		} else {
			// op after new group '(' or after binary operator
			if prev, prevIsOp := prevToken.(*OperatorToken); prevIsOp {
				prevOp := prev.Op
				if !prevOp.Unary {
					// new group or binary operator
					return nextOp, nil
				}
				// unary operator
				if len(prevOp.Raw) > 1 {
					// prev is ++ or --
					if op.Unary || prevOp.RightToLeft {
						// repeat ++ -- or prefix operator can't follow any other operator
						return nil, fmt.Errorf("unexpected operator")
					}
					// take it as binary operator
				} else {
					// prev is unary + - or ! ~
					return nextOp, nil
				}
			}
		}
	}
	return nil, nil
}

func (op *Operator) IsSureSingleOperator() bool {
	return len(op.Raw) == 1 && op.NextMaybe == nil
}

var parenOperator = Operator{
	Raw:      []byte("("),
	Priority: 18,
}

var parenEndOperator = Operator{
	Raw:      []byte(")"),
	Priority: 0,
}

var fnCallOperator = Operator{
	Raw:      []byte("("),
	Priority: 17,
}

var bracketOperator = Operator{
	Raw:      []byte("["),
	Priority: 17,
}

var memberOperator = Operator{
	Raw:      []byte("."), // Member Access
	Priority: 17,
}

var objMemberOperator = Operator{
	Raw:      []byte("->"), // Object Member Access
	Priority: 17,
}

var pipeOperator = Operator{
	Raw:      []byte("|"),
	Priority: 17,
}

var operatorList = []Operator{
	{
		Raw:      []byte("+"), // Addition
		Priority: 11,
		NextMaybe: &Operator{
			Raw:      []byte("+"), // Unary plus
			Unary:    true,
			Priority: 14,
			NextMaybe: &Operator{
				Raw:      []byte("++"), // Postfix Increment
				Priority: 14,
				Unary:    true,
				NextMaybe: &Operator{
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
		NextMaybe: &Operator{
			Raw:      []byte("-"), // Unary negation
			Priority: 14,
			Unary:    true,
			NextMaybe: &Operator{
				Raw:      []byte("--"), // Postfix Decrement
				Priority: 14,
				Unary:    true,
				NextMaybe: &Operator{
					Raw:         []byte("--"), // Prefix Decrement
					Priority:    15,
					Unary:       true,
					RightToLeft: true,
					NextMaybe:   &objMemberOperator,
				},
			},
		},
	}, {
		Raw:      []byte("*"), // Multiplication
		Priority: 12,
		NextMaybe: &Operator{
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
		Raw:      []byte("!"), // Logic Not
		Priority: 14,
		Unary:    true,
		NextMaybe: &Operator{
			Raw:      []byte("!="), // Inequality
			Priority: 8,
		},
	},
	{
		Raw:      []byte(">"), // Greater Than
		Priority: 9,
		NextMaybe: &Operator{
			Raw:      []byte(">="), // Greate Than Or Equal
			Priority: 9,
			NextMaybe: &Operator{
				Raw:      []byte(">>"), // Bitwise Right Shift
				Priority: 10,
			},
		},
	},
	{
		Raw:      []byte("<"), // Less Than
		Priority: 9,
		NextMaybe: &Operator{
			Raw:      []byte("<="), // Less Than Or Equal
			Priority: 9,
			NextMaybe: &Operator{
				Raw:      []byte("<<"), // Bitwise Left Shift
				Priority: 10,
			},
		},
	}, {
		Raw:      []byte("&"), // Bitwise And
		Priority: 7,
		NextMaybe: &Operator{
			Raw:      []byte("&&"), // Logic And
			Priority: 4,
		},
	},
	{
		Raw:      []byte("~"), // Bitwise Not
		Priority: 14,
		Unary:    true,
	},
	{
		Raw:      []byte("^"), // Bitwise XOR
		Priority: 6,
	},
	{
		Raw:      []byte("|"), // Bitwise Or
		Priority: 5,
		NextMaybe: &Operator{
			Raw:      []byte("||"), // Logic Or
			Priority: 3,
		},
	},
}

/**
 * -------------------------
 * Interfaces
 */

type IToken interface {
	Add(bt byte, exp *Expression) (IToken, error)
	Type() TokenType
	End() error
	RawBytes() []byte
}

/**
 * operator token
 */

type OperatorToken struct {
	Op    *Operator
	Index int
}

func (token *OperatorToken) Add(bt byte, exp *Expression) (IToken, error) {
	op := token.Op
	nextIndex := token.Index + 1
	totalByteLen := len(op.Raw)
	// check if is still matched in current operator
	if nextIndex < totalByteLen && op.Raw[nextIndex] == bt {
		token.Index = nextIndex
		return nil, nil
	}
	// check the next operators
	nextOp := op
	for {
		nextOp = nextOp.NextMaybe
		if nextOp == nil {
			break
		}
		if len(nextOp.Raw) > nextIndex && nextOp.Raw[nextIndex] == bt {
			// change op to the next
			if unaryToken, err := nextOp.FixIfUnary(exp.PrevToken); err == nil {
				if unaryToken != nil {
					exp.CurToken = &OperatorToken{
						Op:    unaryToken,
						Index: nextIndex,
					}
				} else {
					exp.CurToken = &OperatorToken{
						Op:    nextOp,
						Index: nextIndex,
					}
				}
			} else {
				return nil, err
			}
			// maybe | => ||
			if exp.LazyPipe != nil {
				exp.LazyPipe = nil
			}
			return nil, nil
		}
	}
	// fix the unary token
	if unaryToken, err := op.FixIfUnary(exp.PrevToken); err == nil {
		if unaryToken != nil {
			exp.CurToken = &OperatorToken{
				Op: unaryToken,
			}
		}
	} else {
		return nil, err
	}
	return AddUnkownTokenByte(bt, exp)
}

func (token *OperatorToken) Type() TokenType {
	return OpType
}

func (token *OperatorToken) End() error {
	if token.Index == len(token.Op.Raw)-1 {
		return nil
	}
	return fmt.Errorf("the operator is not end")
}

func (token *OperatorToken) RawBytes() []byte {
	return token.Op.Raw
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

func (sp *SpaceToken) Add(bt byte, exp *Expression) (IToken, error) {
	if IsSpaceByte(bt) {
		sp.Raw = append(sp.Raw, bt)
		return nil, nil
	}
	return AddUnkownTokenByte(bt, exp)
}

func (sp *SpaceToken) Type() TokenType {
	return SpaceType
}

func (sp *SpaceToken) End() error {
	return nil
}

func (sp *SpaceToken) RawBytes() []byte {
	return sp.Raw
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

func (num *NumberToken) Add(bt byte, exp *Expression) (IToken, error) {
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

func (num *NumberToken) RawBytes() []byte {
	return num.Integer.Raw
}

/**
 * Identifier token, e.g $a abc a123
 */
type IdentifierToken struct {
	IsVariable bool
	Raw        []byte
}

func (id *IdentifierToken) Add(bt byte, exp *Expression) (IToken, error) {
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

func (id *IdentifierToken) RawBytes() []byte {
	return id.Raw
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

func (ss *SingleQuoteStringToken) Add(bt byte, exp *Expression) (IToken, error) {
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

func (ss *SingleQuoteStringToken) RawBytes() []byte {
	return ss.Raw
}

// Double Quote String
type RepExp struct {
	Range []uint
	Exp   *Expression
}
type DoubleQuoteStringToken struct {
	InTranslate bool
	IsEnd       bool
	InExp       bool
	Raw         []byte
	Exps        []RepExp
}

func (ds *DoubleQuoteStringToken) Add(bt byte, exp *Expression) (IToken, error) {
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

func (ds *DoubleQuoteStringToken) RawBytes() []byte {
	return ds.Raw
}

/**
 * Array Literal
 * e.g => [ "a" => 1, "c"]
 */
type ArrayItem struct {
	Key   IToken
	Value IToken
}

type ArrayLiteral struct {
	IsEnd  bool
	HasKey bool
	State  ArrayInState
	Index  int
	CurExp *Expression
	Items  []ArrayItem
	Raw    []byte
}

func (arr *ArrayLiteral) Add(bt byte, exp *Expression) (IToken, error) {

	// first check if is end
	if arr.IsEnd {
		return AddSpaceOrOperatorByte(bt)
	}
	// array state
	curExp := arr.CurExp
	switch arr.State {
	case WaitGreatThan:
		if bt == '>' {
			arr.State = InArrayValue
			arr.CurExp = New()
			return nil, nil
		} else {
			return nil, fmt.Errorf("unexpected token '=%s' in array literal, do you mean '=>'?", string(bt))
		}
	case MaybeArrayKey, InArrayValue:
		isValue := true
		if _, err := curExp.Add(bt, exp); err == nil {
			return nil, nil
		} else {
			if bt == ']' {
				// end of array
				arr.IsEnd = true
			} else if bt == ',' {
				// new array item
				arr.CurExp = New()
			} else if arr.State == MaybeArrayKey && bt == '=' {
				isValue = false
				arr.HasKey = true
				arr.State = WaitGreatThan
			} else {
				return nil, fmt.Errorf("unexpecte token '%s' in array literal", string(bt))
			}
			if err = curExp.Eof(); err != nil {
				// if the byte is end bracket ']'
				// should ignore empty array '[]'
				// or the last array item's end ',', [1,]
				if arr.IsEnd && !arr.HasKey && curExp.IsEmpty() {
					return nil, nil
				} else {
					return nil, err
				}
			}
			lastToken := curExp.Token()
			if isValue {
				// add value
				if arr.HasKey {
					arr.Items[len(arr.Items)-1].Value = lastToken
					arr.HasKey = false
				} else {
					arr.Items = append(arr.Items, ArrayItem{
						Value: lastToken,
					})
				}
			} else {
				// check if key is allowed type token
				curType := lastToken.Type()
				if curType == StrType || curType == NumType || curType == IdentType {
					arr.Items = append(arr.Items, ArrayItem{
						Key: lastToken,
					})
				} else {
					return nil, fmt.Errorf("disallowed key token in array literal")
				}
			}
		}
	}
	return nil, nil
}

func (arr *ArrayLiteral) Type() TokenType {
	return ArrLitType
}

func (arr *ArrayLiteral) End() error {
	return nil
}

func (arr *ArrayLiteral) RawBytes() []byte {
	return arr.Raw
}

/**
 * Function Call
 * e.g => abc(), abc(1+2, "def")
 */
type FunctionCall struct {
	IsEnd  bool
	Name   IToken
	Args   []IToken
	CurArg *Expression
	Raw    []byte
}

func (fn *FunctionCall) Add(bt byte, exp *Expression) (IToken, error) {
	// first check if fn call is end
	if fn.IsEnd {
		return AddSpaceOrOperatorByte(bt)
	}
	// add byte to cur argument
	curArg := fn.CurArg
	if nextToken, err := curArg.Add(bt, curArg); err != nil {
		if bt == ',' {
			// fn(1,
			if err = curArg.Eof(); err == nil {
				fn.Args = append(fn.Args, curArg.Token())
				fn.CurArg = New()
				return nil, nil
			}
		} else if IsOpToken(nextToken, &parenEndOperator) {
			curArg.CurToken = &spaceToken
			// fn(1, 2)
			if err = curArg.Eof(); err == nil {
				fn.IsEnd = true
				fn.Args = append(fn.Args, curArg.Token())
				fn.CurArg = nil
				return nil, nil
			} else if curArg.IsEmpty() {
				// fn(1,)
				fn.IsEnd = true
				return nil, nil
			}
		}
		return nil, err
	}
	return nil, nil
}

func (fn *FunctionCall) Type() TokenType {
	return FuncCallType
}

func (fn *FunctionCall) End() error {
	return nil
}

func (fn *FunctionCall) RawBytes() []byte {
	return fn.Raw
}

/**
*
 */
type PipeFunction struct {
	Name   *IdentifierToken
	Args   []IToken
	CurArg *Expression
	Raw    []byte
}

func (pipe *PipeFunction) Add(bt byte, exp *Expression) (IToken, error) {
	if pipe.CurArg != nil {

	}
	nextToken, err := pipe.Name.Add(bt, exp)
	if err == nil {

	} else if bt == ':' {

	}
	return nextToken, err
}

func (pipe *PipeFunction) Type() TokenType {
	return PipeFuncType
}

func (pipe *PipeFunction) End() error {
	return nil
}

func (pipe *PipeFunction) RawBytes() []byte {
	return pipe.Raw
}

/*
* Object Property
 */
type ObjectProperty struct {
	Object   IToken
	Property *IdentifierToken
	Raw      []byte
}

func (obj *ObjectProperty) Add(bt byte, exp *Expression) (IToken, error) {
	return nil, nil
}

func (obj *ObjectProperty) Type() TokenType {
	return ObjPropType
}

func (obj *ObjectProperty) End() error {
	return nil
}

func (obj *ObjectProperty) RawBytes() []byte {
	return obj.Raw
}

/*
*
* Array Field
* e.g $a.b $a["b"]
 */
type ArrayField struct {
	IsEnd         bool
	Array         IToken
	Field         *IdentifierToken
	ComputedField *Expression
}

func (arr *ArrayField) Add(bt byte, exp *Expression) (IToken, error) {
	// first check if static field
	if arr.Field != nil {
		return arr.Field.Add(bt, exp)
	}
	// then check if the computed field is end
	if arr.IsEnd {
		return AddSpaceOrOperatorByte(bt)
	}
	// add byte to computed field
	dynamicField := arr.ComputedField
	if _, err := dynamicField.Add(bt, dynamicField); err != nil {
		if bt == ']' {
			if err = dynamicField.Eof(); err == nil {
				arr.IsEnd = true
				return nil, nil
			}
		}
		return nil, err
	}
	return nil, nil
}

func (arr *ArrayField) Type() TokenType {
	return ArrFieldType
}

func (arr *ArrayField) End() error {
	return nil
}

func (arr *ArrayField) RawBytes() []byte {
	var buf = bytes.NewBuffer(arr.Array.RawBytes())
	if arr.Field != nil {
		_, _ = buf.Write(arr.Field.RawBytes())
	} else {
		_, _ = buf.Write(arr.ComputedField.RawBytes())
	}
	return buf.Bytes()
}

/**
*
 */
type Ast struct {
	Op    *Operator
	Left  IToken
	Right IToken
}

func (ast *Ast) Add(bt byte, exp *Expression) (IToken, error) {
	return nil, nil
}

func (ast *Ast) Type() TokenType {
	return AstType
}

func (ast *Ast) End() error {
	return nil
}

func (ast *Ast) RawBytes() []byte {
	return []byte{}
}

type Expression struct {
	PrevToken   IToken
	CurToken    IToken
	LazyPipe    *OperatorToken
	ImmediateOp *OperatorToken
	OpStack     []*OperatorToken
	Output      []IToken
}

func New() *Expression {
	return &Expression{
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

/**
 * add operator
 */
func (exp *Expression) addOperator(opToken *OperatorToken) error {
	op := opToken.Op
	opStack := exp.OpStack
	total := len(opStack)
	if total == 0 {
		if op == &parenEndOperator {
			return fmt.Errorf("unexpected operator ')'")
		}
		exp.OpStack = append(opStack, opToken)
	} else {
		if op == &parenEndOperator {
			// pop all operators until meet paren operator (
			isPair := false
			for total != 0 {
				total--
				curOpToken := opStack[total]
				if curOpToken.Op == &parenOperator {
					isPair = true
					break
				}
				// add to output
				if err := exp.addOperatorToOutput(curOpToken); err != nil {
					return err
				}
			}
			if isPair {
				exp.OpStack = opStack[:total]
			} else {
				return fmt.Errorf("unexpect operator ')'")
			}
		} else {
			priority := op.Priority
			index := total
			// pop all operators which priority not less than cur operator
			for index != 0 {
				nextIndex := index - 1
				curOpToken := opStack[nextIndex]
				curOp := curOpToken.Op
				if curOp == &parenOperator || curOp.Priority < priority {
					break
				}
				index = nextIndex
				// add to output
				if err := exp.addOperatorToOutput(curOpToken); err != nil {
					return err
				}
			}
			if index != total {
				exp.OpStack = opStack[:index]
			}
			// add current, here use exp.OpStack
			exp.OpStack = append(exp.OpStack, opToken)
		}
	}
	return nil
}

func (exp *Expression) addOperatorToOutput(opToken *OperatorToken) error {
	op := opToken.Op
	output := exp.Output
	total := len(output)
	if op.Unary {
		// unary operator
		if total == 0 {
			return fmt.Errorf("unexpected unary operator")
		}
		lastToken := output[total-1]
		if len(op.Raw) > 1 {
			// postfix or prefix ++ --
			if lastToken.Type() != IdentType {
				return fmt.Errorf("invalid left-hand side in assignment")
			}
		}
		ast := &Ast{
			Op: op,
		}
		if op.RightToLeft {
			ast.Right = lastToken
		} else {
			ast.Left = lastToken
		}
		output[total-1] = ast
	} else {
		if total < 2 {
			return fmt.Errorf("unexpected operator: %s", string(op.Raw))
		}
		left, right := output[total-2], output[total-1]
		var ast IToken
		if op == &fnCallOperator {
			// function call
			if fn, isFn := right.(*FunctionCall); isFn {
				fn.Name = left
				ast = right
			} else {
				return fmt.Errorf("wrong function call")
			}
		} else if op == &memberOperator {
			// array static field
			if ident, isIdent := right.(*IdentifierToken); isIdent {
				ast = &ArrayField{
					Array: left,
					Field: ident,
				}
			} else {
				return fmt.Errorf("unexpect array field")
			}
		} else if op == &bracketOperator {
			// array computed field
			if arr, isArr := right.(*ArrayField); isArr {
				arr.Array = left
				ast = arr
			} else {
				return fmt.Errorf("unexpected array ")
			}
		} else if op == &pipeOperator {
			// pipe function
			if fn, isFn := right.(*PipeFunction); isFn {
				// reset the args
				fn.Args = append([]IToken{left}, fn.Args...)
				// set the fn as AST
				ast = fn
			} else {
				return fmt.Errorf("unexpected pipe function")
			}
		} else if op == &objMemberOperator {
			// object property
			if ident, isIdent := right.(*IdentifierToken); isIdent {
				ast = &ObjectProperty{
					Object:   left,
					Property: ident,
				}
			} else {
				return fmt.Errorf("unexpected object member")
			}
		} else {
			ast = &Ast{
				Op:    op,
				Left:  left,
				Right: right,
			}
		}
		output[total-2] = ast
		exp.Output = exp.Output[:total-1]
	}
	return nil
}

func (exp *Expression) IsEmpty() bool {
	return len(exp.Output) == 0
}

func (exp *Expression) Add(bt byte, _ *Expression) (IToken, error) {
	var curToken = exp.CurToken
	if nextToken, err := curToken.Add(bt, exp); err == nil {
		// change current token to next
		// if next token is not nil
		if nextToken != nil {
			isLazyPipeAdded := false
			// special operators
			if opToken, isOp := nextToken.(*OperatorToken); isOp {
				op := opToken.Op
				if op == &parenOperator {
					// '('
					prevType := exp.getPrevToken().Type()
					if prevType == IdentType || prevType == FuncCallType || prevType == ObjPropType || prevType == ArrFieldType {
						// check if function call
						nextToken = &FunctionCall{
							CurArg: New(),
						}
						// add function call operator
						fnOpToken := OperatorToken{
							Op: &fnCallOperator,
						}
						if err = exp.addOperator(&fnOpToken); err != nil {
							return &fnOpToken, err
						}
						exp.PrevToken = &fnOpToken
					}
				} else if op == &bracketOperator {
					// '[', if before is '(' or the beginning, take it as a array literal type
					prevToken := exp.getPrevToken()
					if IsOpToken(prevToken, &parenOperator) || prevToken.Type() == SpaceType {
						nextToken = &ArrayLiteral{
							CurExp: New(),
							State:  MaybeArrayKey,
						}
					} else {
						// '[' computed array field
						nextToken = &ArrayField{
							ComputedField: New(),
						}
						// add the '[' operator
						bracketOpToken := OperatorToken{
							Op: &bracketOperator,
						}
						if err = exp.addOperator(&bracketOpToken); err != nil {
							return opToken, err
						}
						exp.PrevToken = &bracketOpToken
					}
				} else if IsPipeOpToken(opToken) {
					// | operator should judge lazy
					exp.LazyPipe = opToken
				} else if op.IsSureSingleOperator() {
					fmt.Println("IMMEDIATE====>")
					fmt.Printf("op====> %s", string(op.Raw))
					fmt.Println()
					exp.ImmediateOp = opToken

					spew.Dump(exp)
				}
			} else if exp.LazyPipe != nil && nextToken.Type() != SpaceType {
				if ident, isIdent := nextToken.(*IdentifierToken); isIdent && !ident.IsVariable {
					// pipe function, a|count
					nextToken = &PipeFunction{
						Name: ident,
					}
					// replace the last op from (bitwise |) to (pipe |)
					exp.LazyPipe = &OperatorToken{
						Op: &pipeOperator,
					}
				}
				// add lazy pipe to operator stack
				if err = exp.addOperator(exp.LazyPipe); err != nil {
					return exp.LazyPipe, err
				}
				// set lazy pipe has been added to op stack
				isLazyPipeAdded = true
				// reset lazy pipe
				exp.LazyPipe = nil
			}
			// if not space type, should add the current token into output or op stack
			if curToken.Type() != SpaceType {
				// check if cur token is operator or normal token
				if opToken, isOp := curToken.(*OperatorToken); isOp {
					op := opToken.Op
					// check if repeated non unary operators
					if !op.Unary {
						if prevOpToken, prevIsOp := exp.PrevToken.(*OperatorToken); prevIsOp && !prevOpToken.Op.Unary {
							return nil, fmt.Errorf("Unexpected token '%s'", string(op.Raw))
						}
					}
					// only not lazy pipe and not immediate op added to the op stack
					if exp.ImmediateOp != nil {
						exp.ImmediateOp = nil
					} else if exp.LazyPipe == nil && !isLazyPipeAdded {
						if err = exp.addOperator(opToken); err != nil {
							return opToken, err
						}
					}
				} else {
					// other token should added to output
					exp.Output = append(exp.Output, curToken)
				}
				// set prev token
				exp.PrevToken = curToken
			}
			// immediate op
			if exp.ImmediateOp != nil {
				curOpToken := exp.ImmediateOp
				if err = exp.addOperator(curOpToken); err != nil {
					return curOpToken, err
				}

			}
			// set current token
			exp.CurToken = nextToken
		}
		return exp.CurToken, nil
	} else {
		return nil, err
	}
}

func (exp *Expression) Type() TokenType {
	return ExpType
}

func (exp *Expression) End() error {
	return nil
}

func (exp *Expression) RawBytes() []byte {
	return []byte{}
}

/**
* End of the expression
 */
func (exp *Expression) Eof() error {
	fmt.Println("执行===>eof===>output===>")
	spew.Dump(exp)
	// add a space make sure end of the expression
	if !IsSpaceToken(exp.CurToken) {
		if _, err := exp.Add(BYTE_SPACE, exp); err != nil {
			return err
		}
	}
	// check if the last token is end
	if err := exp.CurToken.End(); err != nil {
		return err
	}
	// output all the operators still left in op stack
	opStack := exp.OpStack
	total := len(opStack)
	if total > 0 {
		for total != 0 {
			total--
			curOpToken := opStack[total]
			if curOpToken.Op == &parenOperator {
				// unclosed paren operator
				return fmt.Errorf("unclosed operator '('")
			}
			if err := exp.addOperatorToOutput(curOpToken); err != nil {
				return err
			}
		}
	}
	// translate the output to AST
	if len(exp.Output) != 1 {
		return fmt.Errorf("wrong expression")
	}

	return nil
}

func (exp *Expression) Token() IToken {
	return exp.Output[0]
}

/**
* Parse string to AST
 */
func (exp *Expression) Parse(str string) (*Ast, error) {
	for i := 0; i < len(str); i++ {
		if _, err := exp.Add(str[i], exp); err != nil {
			return nil, err
		}
	}
	if err := exp.Eof(); err != nil {
		return nil, err
	}
	return &Ast{}, nil
}
