/**
 * The lexer for parsing fet template
 */
package lexer

import (
	"bytes"
	"fmt"
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

var KEYWORDS = [][]byte{
	[]byte("true"),
	[]byte("false"),
	[]byte("break"),
	[]byte("continue"),
	[]byte("if"),
	[]byte("elseif"),
	[]byte("for"),
	[]byte("foreach"),
	[]byte("as"),
}

const (
	HexBase     NumberBase = 16
	OctalBase   NumberBase = 8
	BinaryBase  NumberBase = 2
	DecimalBase NumberBase = 10
)

const (
	MaybeArrayKey ArrayInState = iota
	InArrayValue
)

/**
 * same bytes
 */
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

func IsOpToken(token IToken, op *Operator) bool {
	if opToken, isOp := token.(*OperatorToken); isOp && opToken.Op == op {
		return true
	}
	return false
}

func IsBinaryOrParenOpToken(token IToken) bool {
	if opToken, isOp := token.(*OperatorToken); isOp && (opToken.Op == &parenOperator || !opToken.Op.Unary) {
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

func IsEqualOpToken(token IToken) bool {
	if opToken, isOp := token.(*OperatorToken); isOp && opToken.Op == &equalOperator && opToken.Index == 0 {
		return true
	}
	return false
}

func AddSpaceOrOperatorByte(bt byte, args ...[]byte) (IToken, error) {
	if IsSpaceByte(bt) {
		return &SpaceToken{
			Raw: []byte{bt},
		}, nil
	}
	if len(args) > 0 {
		dis := args[0]
		for _, opByte := range dis {
			if bt == opByte {
				return nil, fmt.Errorf("disallowed operator token '%s'", string(opByte))
			}
		}
	}
	// array literal or computed field
	if bt == '[' {
		return &OperatorToken{
			Op: &bracketOperator,
		}, nil
	}
	if bt == ']' {
		return &OperatorToken{
			Op: &bracketEndOperator,
		}, nil
	}
	// paren
	if bt == '(' {
		return &OperatorToken{
			Op: &parenOperator,
		}, nil
	}
	// paren end
	if bt == ')' {
		return &OperatorToken{
			Op: &parenEndOperator,
		}, nil
	}
	// member
	if bt == '.' {
		return &OperatorToken{
			Op: &memberOperator,
		}, nil
	}
	// equal
	if bt == '=' {
		return &OperatorToken{
			Op: &equalOperator,
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
	if token, err := AddSpaceOrOperatorByte(bt, exp.disOp()); err == nil {
		return token, nil
	}
	// double quote
	if bt == '"' {
		return DoubleQuoteString(), nil
	}
	// single quote
	if bt == '\'' {
		return &SingleQuoteStringToken{}, nil
	}
	// try number token
	num := Number()
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

func (op *Operator) IsPipe() bool {
	return len(op.Raw) == 1 && op.Raw[0] == '|'
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

var bracketEndOperator = Operator{
	Raw:      []byte("]"),
	Priority: 0,
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

var equalOperator = Operator{
	Raw:      []byte("=="), // Equal
	Priority: 8,
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
	},
	{
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
	Index int
	Op    *Operator
}

func (token *OperatorToken) Add(bt byte, exp *Expression) (IToken, error) {
	op := token.Op
	nextIndex := token.Index + 1
	totalByteLen := len(op.Raw)
	// check if is still matched in current operator
	if nextIndex < totalByteLen {
		if op.Raw[nextIndex] == bt {
			token.Index = nextIndex
			return nil, nil
		} else if op.NextMaybe == nil {
			return token, fmt.Errorf("unexpected operator '%s%s', do you mean '%s'?", string(op.Raw[:nextIndex]), string(bt), string(op.Raw))
		}
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
					token = &OperatorToken{
						Op:    unaryToken,
						Index: nextIndex,
					}
				} else {
					token = &OperatorToken{
						Op:    nextOp,
						Index: nextIndex,
					}
				}
				exp.CurToken = token
				return nil, nil
			} else {
				return nil, err
			}
		}
	}
	// check if token is end
	if err := token.End(); err != nil {
		return token, err
	}
	// fix the unary token
	if unaryToken, err := op.FixIfUnary(exp.PrevToken); err == nil {
		if unaryToken != nil {
			token = &OperatorToken{
				Op: unaryToken,
			}
			exp.CurToken = token
		}
		return AddUnkownTokenByte(bt, exp)
	} else {
		return nil, err
	}
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

func Number() *NumberToken {
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
							} else if bt == '8' || bt == '9' {
								// wrong octal number
								return nil, fmt.Errorf("wrong octal number")
							} else {
								return AddSpaceOrOperatorByte(bt, exp.disOp())
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
				return AddSpaceOrOperatorByte(bt, exp.disOp())
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
	IsKeyword bool
	IsVar     bool
	Raw       []byte
}

func (id *IdentifierToken) Add(bt byte, exp *Expression) (IToken, error) {
	if len(id.Raw) == 0 {
		if bt == '$' {
			id.IsVar = true
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
			if len(id.Raw) == 1 && id.IsVar {
				return nil, fmt.Errorf("wrong variable $")
			}
			// next only allow space or operator
			return AddSpaceOrOperatorByte(bt, exp.disOp())
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
 * DoubleQuote: "abc`$a`"
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
		return AddSpaceOrOperatorByte(bt, exp.disOp())
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
	buf := bytes.Buffer{}
	buf.WriteByte('\'')
	_, _ = buf.Write(ss.Raw)
	buf.WriteByte('\'')
	return buf.Bytes()
}

// Double Quote String
func DoubleQuoteString() *DoubleQuoteStringToken {
	return &DoubleQuoteStringToken{
		CurRaw: []byte{},
	}
}

type StringVariable struct {
	Index int
	Var   *Expression
}

type DoubleQuoteStringToken struct {
	InTranslate bool
	IsEnd       bool
	CurRawLen   int
	CurVar      *Expression
	CurRaw      []byte
	RawString   [][]byte
	VarString   []StringVariable
}

func (ds *DoubleQuoteStringToken) Add(bt byte, exp *Expression) (IToken, error) {
	// check if is end
	if ds.IsEnd {
		return AddSpaceOrOperatorByte(bt, exp.disOp())
	}
	// check if in variable
	if ds.CurVar != nil {
		curVar := ds.CurVar
		if bt == '"' {
			// not a string variable
			ds.RawString = append(ds.RawString, ds.CurRaw)
			ds.CurVar = nil
			ds.CurRaw = nil
			ds.IsEnd = true
		} else if bt == '`' {
			if err := curVar.Eof(); err != nil {
				return curVar, err
			}
			// add the raw string first
			if ds.CurRawLen > 0 {
				ds.RawString = append(ds.RawString, ds.CurRaw[:ds.CurRawLen])
			}
			// add variable
			ds.VarString = append(ds.VarString, StringVariable{
				Var:   curVar,
				Index: len(ds.RawString),
			})
			// reset cur var and raw
			ds.CurVar = nil
			ds.CurRaw = []byte{}
		} else {
			// add byte to variable
			if errToken, err := curVar.Add(bt, curVar); err != nil {
				return errToken, err
			}
			// also added to cur raw
			ds.CurRaw = append(ds.CurRaw, bt)
		}
	} else {
		if ds.InTranslate {
			ds.InTranslate = false
			ds.CurRaw = append(ds.CurRaw, bt)
		} else if bt == '\\' {
			// translate
			ds.InTranslate = true
		} else if bt == '`' {
			// expression
			ds.CurVar = New()
			ds.CurRawLen = len(ds.CurRaw)
			ds.CurRaw = append(ds.CurRaw, bt)
		} else if bt == '"' {
			// end '
			if len(ds.CurRaw) > 0 {
				ds.RawString = append(ds.RawString, ds.CurRaw)
			}
			ds.IsEnd = true
			ds.CurRaw = nil
		} else {
			// add byte
			ds.CurRaw = append(ds.CurRaw, bt)
		}
	}
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
	buf := bytes.Buffer{}
	buf.WriteByte('"')
	if len(ds.VarString) > 0 {
		raws := ds.RawString
		startIndex := 0
		// add raw and var bytes
		for _, varStr := range ds.VarString {
			index := varStr.Index
			for startIndex < index {
				buf.WriteByte('`')
				_, _ = buf.Write(raws[startIndex])
				buf.WriteByte('`')
				startIndex++
			}
			_, _ = buf.Write(varStr.Var.RawBytes())
		}
		// add the left raw bytes
		for startIndex < len(raws) {
			_, _ = buf.Write(raws[startIndex])
			startIndex++
		}
	} else {
		if len(ds.RawString) == 1 {
			_, _ = buf.Write(ds.RawString[0])
		}
	}
	buf.WriteByte('"')
	return buf.Bytes()
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

func (arr *ArrayLiteral) SetCurExp() {
	// disable operator ']'
	// so the parser can process it in the feature err not nil
	arr.CurExp = New(&ExpOpts{
		DisAllowedOp: []byte{']'},
	})
}

func (arr *ArrayLiteral) Add(bt byte, exp *Expression) (IToken, error) {
	// first check if is the array is end
	if arr.IsEnd {
		return AddSpaceOrOperatorByte(bt, exp.disOp())
	}
	// array state
	curExp := arr.CurExp
	if _, err := curExp.Add(bt, curExp); err == nil {

		return nil, nil
	} else {
		isValue := true
		if bt == ']' {
			// end of array
			arr.IsEnd = true
		} else if bt == ',' {
			// new array item
			arr.SetCurExp()
		} else if arr.State == MaybeArrayKey && IsEqualOpToken(curExp.CurToken) && bt == '>' {
			// when '=>'
			curExp.CurToken = &spaceToken
			isValue = false
			arr.HasKey = true
			arr.State = InArrayValue
			arr.SetCurExp()
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
			arr.State = MaybeArrayKey
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
	// add byte to cur argument
	curArg := fn.CurArg
	if nextToken, err := curArg.Add(bt, curArg); err != nil {
		if bt == ',' {
			// fn(1,
			if err = curArg.Eof(); err != nil {
				return curArg, err
			}
			fn.Args = append(fn.Args, curArg.Token())
			fn.CurArg = New()
		} else if IsOpToken(nextToken, &parenEndOperator) {
			curArg.CurToken = &spaceToken
			fn.IsEnd = true
			// fn(1, 2)
			if err = curArg.Eof(); err != nil {
				if curArg.IsEmpty() {
					// fn(1,)
					return AddSpaceOrOperatorByte(bt, exp.disOp())
				}
				return curArg, err
			}
			fn.Args = append(fn.Args, curArg.Token())
			fn.CurArg = nil
			// when end should return new token
			return AddSpaceOrOperatorByte(bt, exp.disOp())
		} else {
			return nil, err
		}
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
	CurArg *Expression
	Name   *IdentifierToken
	Args   []IToken
	Raw    []byte
}

func (pipe *PipeFunction) Add(bt byte, exp *Expression) (IToken, error) {
	curArg := pipe.CurArg
	if curArg != nil {
		if errToken, err := curArg.Add(bt, curArg); err == nil {
			opStack := curArg.OpStack
			if len(opStack) == 1 {
				opToken := opStack[0]
				isPipe := opToken.Op == &pipeOperator
				if isPipe || opToken.Op.Priority < pipeOperator.Priority {
					// save curArg's token to exp's cur token
					curArgToken := curArg.CurToken
					// remove the operator from cur arg expression
					curArg.OpStack = opStack[:0]
					// ignore cur arg's cur token
					curArg.CurToken = &spaceToken
					if err = curArg.Eof(); err != nil {
						return curArg, err
					}
					pipe.Args = append(pipe.Args, curArg.Token())
					pipe.CurArg = nil
					// set prev token
					if isPipe {
						exp.Output = append(exp.Output, pipe)
						if nextToken, err := exp.addConvertOperator(&pipeOperator); err != nil {
							return nextToken, err
						}
					} else {
						exp.Output = append(exp.Output, pipe)
						if err = exp.addOperator(opToken); err != nil {
							return opToken, err
						}
						exp.PrevToken = opToken
					}
					exp.CurToken = curArgToken
					return nil, nil
				}
			}
			return nil, nil
		} else {
			if bt == ':' {
				if err = curArg.Eof(); err == nil {
					pipe.Args = append(pipe.Args, curArg.Token())
					pipe.CurArg = New()
					return nil, nil
				}
			}
			return errToken, err
		}
	}
	// still in function name
	if errToken, err := pipe.Name.Add(bt, exp); err != nil {
		if bt == ':' {
			pipe.CurArg = New()
		} else {
			return errToken, nil
		}
	}
	return nil, nil
}

func (pipe *PipeFunction) Type() TokenType {
	return PipeFuncType
}

func (pipe *PipeFunction) End() error {
	curArg := pipe.CurArg
	if curArg != nil {
		if err := curArg.Eof(); err != nil {
			return err
		}
		pipe.Args = append(pipe.Args, curArg.Token())
		pipe.CurArg = nil
	}
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
		return AddSpaceOrOperatorByte(bt, exp.disOp())
	}
	// add byte to computed field
	computedField := arr.ComputedField
	if _, err := computedField.Add(bt, computedField); err != nil {
		if bt == ']' {
			if err = computedField.Eof(); err == nil {
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

type ExpOpts struct {
	DisAllowedOp []byte
}

type Expression struct {
	PrevToken   IToken
	CurToken    IToken
	LazyPipe    *OperatorToken
	ImmediateOp *OperatorToken
	Options     *ExpOpts
	OpStack     []*OperatorToken
	Output      []IToken
}

func New(args ...*ExpOpts) *Expression {
	exp := Expression{
		CurToken:  &spaceToken,
		PrevToken: &spaceToken,
	}
	if len(args) > 0 {
		exp.Options = args[0]
	}
	return &exp
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

func (exp *Expression) addConvertOperator(op *Operator) (IToken, error) {
	opToken := OperatorToken{
		Op: op,
	}
	if err := exp.addOperator(&opToken); err != nil {
		return &opToken, err
	}
	exp.PrevToken = &opToken
	return nil, nil
}

func (exp *Expression) disOp() []byte {
	if exp.Options != nil {
		return exp.Options.DisAllowedOp
	}
	return []byte{}
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
		// if next token is not nil
		// current token is end
		if nextToken != nil {
			// first check if cur token is end
			if err = curToken.End(); err != nil {
				return curToken, err
			}
			// first do with lazy pipe
			if exp.LazyPipe != nil {
				// cur token is space token, and next token must be non space token
				if ident, isIdent := nextToken.(*IdentifierToken); isIdent && !ident.IsVar {
					// pipe function
					nextToken = &PipeFunction{
						Name: ident,
					}
					// add the '|' pipe operator
					if errToken, err := exp.addConvertOperator(&pipeOperator); err != nil {
						return errToken, err
					}
					exp.CurToken = nextToken
					exp.LazyPipe = nil
					return nextToken, nil
				} else {
					// add bitwise |
					if err := exp.addOperator(exp.LazyPipe); err != nil {
						return exp.LazyPipe, err
					}
					// set cur token
					exp.CurToken = nextToken
					// jump next steps
					return nextToken, nil
				}
			} else {
				// check other condition
				if opToken, isOp := curToken.(*OperatorToken); isOp {
					// operators
					op := opToken.Op
					isConvOp := false
					if op == &parenOperator {
						// operator (
						prevType := exp.PrevToken.Type()
						if prevType == IdentType || prevType == FuncCallType || prevType == ObjPropType || prevType == ArrFieldType {
							// check if function call
							nextToken = &FunctionCall{
								CurArg: New(),
							}
							// add function call operator
							if errToken, err := exp.addConvertOperator(&fnCallOperator); err != nil {
								return errToken, err
							}
							isConvOp = true
						}
					} else if op == &bracketOperator {
						// '[', if before is binary operator or paren or the beginning
						// take it as an array literal type
						prevToken := exp.PrevToken
						if IsBinaryOrParenOpToken(prevToken) || IsSpaceToken(prevToken) {
							arrLit := &ArrayLiteral{
								State: MaybeArrayKey,
							}
							arrLit.SetCurExp()
							nextToken = arrLit
						} else {
							// '[' computed array field
							nextToken = &ArrayField{
								ComputedField: New(&ExpOpts{
									DisAllowedOp: []byte{']'},
								}),
							}
							// add the '[' operator
							if errToken, err := exp.addConvertOperator(&bracketOperator); err != nil {
								return errToken, err
							}
						}
						// always converted op
						isConvOp = true
					} else if op.IsPipe() {
						// operator |
						nextType := nextToken.Type()
						if nextType == IdentType {
							ident := nextToken.(*IdentifierToken)
							if !ident.IsVar {
								// pipe function
								nextToken = &PipeFunction{
									Name: ident,
								}
								// add the '|' operator
								if errToken, err := exp.addConvertOperator(&pipeOperator); err != nil {
									return errToken, err
								}
								exp.CurToken = nextToken
								return nextToken, nil
							}
						} else if nextType == SpaceType {
							// space token need lazy check
							exp.LazyPipe = opToken
							// reset prev and cur token
							exp.PrevToken = curToken
							exp.CurToken = nextToken
							return nextToken, nil
						}
					}
					// if is convert op
					if isConvOp {
						// reset cur token, prev token is set by convert token
						exp.CurToken = nextToken
						return nextToken.Add(bt, exp)
					}
					// check if repeated non unary operators
					if !op.Unary && op != &parenOperator && op != &bracketOperator {
						if prevOpToken, prevIsOp := exp.PrevToken.(*OperatorToken); prevIsOp && !prevOpToken.Op.Unary {
							return nil, fmt.Errorf("unexpected operator token '%s'", string(op.Raw))
						}
					}
					// bracket end operator ]
					if op == &bracketEndOperator {
						return opToken, fmt.Errorf("unexpected operator ']'")
					}
					// add operator to op stack
					if err = exp.addOperator(opToken); err != nil {
						return opToken, err
					}
					// reset prev and cur token
					exp.PrevToken = curToken
					exp.CurToken = nextToken
				} else {
					// output
					if !IsSpaceToken(curToken) {
						// ignore space token
						exp.Output = append(exp.Output, curToken)
						// no space token need reset prev token
						exp.PrevToken = curToken
					}
					exp.CurToken = nextToken
				}
			}
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
	return exp.Token().RawBytes()
}

/**
* End of the expression
 */
func (exp *Expression) Eof() error {
	curToken := exp.CurToken
	if !IsSpaceToken(curToken) {
		// check if the last token is end
		if err := curToken.End(); err != nil {
			return err
		}
		// add a space make sure end of the expression
		if _, err := exp.Add(BYTE_SPACE, exp); err != nil {
			return err
		}
		// pipe function need add to output
		if curToken.Type() == PipeFuncType {
			exp.Output = append(exp.Output, curToken)
		}
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
func (exp *Expression) Parse(str string) (IToken, error) {
	for i := 0; i < len(str); i++ {
		if errToken, err := exp.Add(str[i], exp); err != nil {
			return errToken, err
		}
	}
	if err := exp.Eof(); err != nil {
		return exp.CurToken, err
	}
	return exp.Token(), nil
}
