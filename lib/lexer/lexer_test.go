package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func parseAst(str string) (IToken, error) {
	return New().Parse(str)
}

func isTokenType(str string, tokenType TokenType) bool {
	if token, err := parseAst(str); err == nil {
		return token.Type() == tokenType
	}
	return false
}

func isWrongTokenType(str string, tokenType TokenType) bool {
	exp := New()
	if _, err := exp.Parse(str); err != nil && exp.CurToken.Type() == tokenType {
		return true
	}
	return false
}

func TestNumberToken(t *testing.T) {
	var isNumber = func(str string) bool {
		return isTokenType(str, NumType)
	}
	var isWrongNumber = func(str string) bool {
		return isWrongTokenType(str, NumType)
	}
	// ----------integer--------
	assert.True(t, isNumber("0"))
	assert.True(t, isNumber("9"))
	assert.True(t, isNumber("3"))
	assert.True(t, isNumber("520"))
	assert.True(t, isNumber("5_20"))
	assert.True(t, isNumber("52_0"))
	assert.True(t, isNumber("5_2_0"))
	// wrong integer
	assert.True(t, isWrongNumber("5__2_0"))
	assert.True(t, isWrongNumber("5_2_0_"))
	// float
	assert.True(t, isNumber("0.3"))
	assert.True(t, isNumber("9.275_2"))
	assert.True(t, isNumber("3.12"))
	assert.True(t, isNumber("520.5_3"))
	assert.True(t, isNumber("5_20.1_2_3"))
	assert.True(t, isNumber("52_0.87"))
	assert.True(t, isNumber("5_2_0.12_3"))
	assert.True(t, isWrongNumber("1_3._1"))
	assert.True(t, isWrongNumber("1_3.1__1 "))
	assert.True(t, isWrongNumber("1_3.1_"))
	// -------------------
	// exponent
	assert.True(t, isNumber("0.3e1"))
	assert.True(t, isNumber("9.275_2E-1"))
	assert.True(t, isNumber("3.1_2e+00_1"))
	assert.True(t, isNumber("12E10"))
	assert.True(t, isNumber("1_2E1_8"))
	// wrong exponent
	assert.True(t, isWrongNumber("1_e1"))
	assert.True(t, isWrongNumber("1_3.e1"))
	assert.True(t, isWrongNumber("1_3.1_e1"))
	assert.True(t, isWrongNumber("1_3.e_1"))
	assert.True(t, isWrongNumber("1_3.e-_1"))
	assert.True(t, isWrongNumber("1_3.e+_1"))
	assert.True(t, isWrongNumber("1_3.e+1_"))
	// binary number
	assert.True(t, isNumber("0b1"))
	assert.True(t, isNumber("0b101"))
	assert.True(t, isNumber("0b1_01"))
	assert.True(t, isNumber("0b1_0_1"))
	assert.True(t, isNumber("0B1"))
	assert.True(t, isNumber("0B101"))
	// octal number
	assert.True(t, isNumber("01"))
	assert.True(t, isNumber("001"))
	assert.True(t, isNumber("0_01"))
	assert.True(t, isNumber("0_0_1"))
	assert.True(t, isNumber("0o1067"))
	assert.True(t, isNumber("0o1_06_7"))
	assert.True(t, isNumber("0o755"))
	assert.True(t, isNumber("0O362"))
	assert.True(t, isNumber("0O577"))
	// hex number
	assert.True(t, isNumber("0x19f"))
	assert.True(t, isNumber("0xff_f"))
	assert.True(t, isNumber("0x1a_F_c"))
	assert.True(t, isNumber("0x3_e_1"))
	assert.True(t, isNumber("0X1_9_F"))
	assert.True(t, isNumber("0X302f"))
}

func TestIdentifierToken(t *testing.T) {
	var isIdent = func(str string) bool {
		return isTokenType(str, IdentType)
	}
	var isWrongIdent = func(str string) bool {
		return isWrongTokenType(str, IdentType)
	}
	// identifer
	assert.True(t, isIdent("$a"))
	assert.True(t, isIdent("$0"))
	assert.True(t, isIdent("$_"))
	assert.True(t, isIdent("_a"))
	assert.True(t, isIdent("_"))
	assert.True(t, isIdent("a"))
	assert.True(t, isIdent("a1"))
	assert.True(t, isIdent("z_"))
	assert.True(t, isIdent("Z_1"))
	assert.True(t, isWrongIdent("$"))
	assert.True(t, isWrongIdent("0a"))
	assert.True(t, isWrongIdent("*a"))
}

func TestDoubleStringToken(t *testing.T) {
	var isString = func(str string) bool {
		return isTokenType(str, StrType)
	}
	var isWrongString = func(str string) bool {
		return isWrongTokenType(str, StrType)
	}
	// double string
	assert.True(t, isString("\"abc\""))
	assert.True(t, isWrongString("\"abc"))
}

func TestPipeFunctionCall(t *testing.T) {
	var isPipe = func(str string) bool {
		return isTokenType(str, PipeFuncType)
	}
	// double string
	assert.True(t, isPipe("123|min"))
}

func TestMain(t *testing.T) {
	exp := New()
	_, err := exp.Parse("123|min|max")
	assert.Error(t, err)
	assert.Nil(t, err)
}
