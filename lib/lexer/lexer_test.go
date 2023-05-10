package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func parseToken(str string, token IToken) (IToken, error) {
	var t IToken
	var err error
	for i := 0; i < len(str); i++ {
		if t, err = token.Add(str[i], nil); err != nil {
			return t, err
		}
	}
	return t, err
}

func parseTokenError(str string, token IToken) error {
	_, err := parseToken(str, token)
	return err
}

func parseTokenNext(str string, token IToken) IToken {
	if nextToken, err := parseToken(str, token); err == nil {
		return nextToken
	}
	return nil
}
func TestNumberToken(t *testing.T) {
	// ----------integer--------
	assert.Nil(t, parseTokenError("0", Number()))
	assert.Nil(t, parseTokenError("9", Number()))
	assert.Nil(t, parseTokenError("3", Number()))
	assert.Nil(t, parseTokenError("520", Number()))
	assert.Nil(t, parseTokenError("5_20", Number()))
	assert.Nil(t, parseTokenError("52_0", Number()))
	assert.Nil(t, parseTokenError("5_2_0", Number()))
	assert.Error(t, parseTokenError("5__2_0", Number()))
	assert.Error(t, parseTokenError("5_2_0_ ", Number()))
	// float
	assert.Nil(t, parseTokenError("0.3", Number()))
	assert.Nil(t, parseTokenError("9.275_2", Number()))
	assert.Nil(t, parseTokenError("3.12", Number()))
	assert.Nil(t, parseTokenError("520.5_3", Number()))
	assert.Nil(t, parseTokenError("5_20.1_2_3", Number()))
	assert.Nil(t, parseTokenError("52_0.87", Number()))
	assert.Nil(t, parseTokenError("5_2_0.12_3", Number()))
	assert.Error(t, parseTokenError("1_3._1", Number()))
	assert.Error(t, parseTokenError("1_3.1__1 ", Number()))
	assert.Error(t, parseTokenError("1_3.1_ ", Number()))
	// exponent
	assert.Nil(t, parseTokenError("0.3e1", Number()))
	assert.Nil(t, parseTokenError("9.275_2E-1", Number()))
	assert.Nil(t, parseTokenError("3.1_2e+00_1", Number()))
	assert.Nil(t, parseTokenError("12E10", Number()))
	assert.Nil(t, parseTokenError("1_2E1_8", Number()))
	assert.Error(t, parseTokenError("1_e1", Number()))
	assert.Error(t, parseTokenError("1_3.e1", Number()))
	assert.Error(t, parseTokenError("1_3.1_e1", Number()))
	assert.Error(t, parseTokenError("1_3.e_1", Number()))
	assert.Error(t, parseTokenError("1_3.e-_1", Number()))
	assert.Error(t, parseTokenError("1_3.e+_1", Number()))
	assert.Error(t, parseTokenError("1_3.e+1_ ", Number()))
	// binary number
	assert.Nil(t, parseTokenError("0b1", Number()))
	assert.Nil(t, parseTokenError("0b101", Number()))
	assert.Nil(t, parseTokenError("0b1_01", Number()))
	assert.Nil(t, parseTokenError("0b1_0_1", Number()))
	assert.Nil(t, parseTokenError("0B1", Number()))
	assert.Nil(t, parseTokenError("0B101", Number()))
	// octal number
	assert.Nil(t, parseTokenError("01", Number()))
	assert.Nil(t, parseTokenError("001", Number()))
	assert.Nil(t, parseTokenError("0_01", Number()))
	assert.Nil(t, parseTokenError("0_0_1", Number()))
	assert.Nil(t, parseTokenError("0o1067", Number()))
	assert.Nil(t, parseTokenError("0o1_06_7", Number()))
	assert.Nil(t, parseTokenError("0o755", Number()))
	assert.Nil(t, parseTokenError("0O362", Number()))
	assert.Nil(t, parseTokenError("0O577", Number()))
	// hex number
	assert.Nil(t, parseTokenError("0x19f", Number()))
	assert.Nil(t, parseTokenError("0xff_f", Number()))
	assert.Nil(t, parseTokenError("0x1a_F_c", Number()))
	assert.Nil(t, parseTokenError("0x3_e_1", Number()))
	assert.Nil(t, parseTokenError("0X1_9_F", Number()))
	assert.Nil(t, parseTokenError("0X302f", Number()))
}

func TestIdentifierToken(t *testing.T) {
	// identifer
	assert.Nil(t, parseTokenError("$a", &IdentifierToken{}))
	assert.Nil(t, parseTokenError("$0", &IdentifierToken{}))
	assert.Nil(t, parseTokenError("$_", &IdentifierToken{}))
	assert.Nil(t, parseTokenError("_a", &IdentifierToken{}))
	assert.Nil(t, parseTokenError("_", &IdentifierToken{}))
	assert.Nil(t, parseTokenError("a", &IdentifierToken{}))
	assert.Nil(t, parseTokenError("a1", &IdentifierToken{}))
	assert.Nil(t, parseTokenError("z_", &IdentifierToken{}))
	assert.Nil(t, parseTokenError("Z_1", &IdentifierToken{}))
	assert.Error(t, parseTokenError("$ ", &IdentifierToken{}))
	assert.Error(t, parseTokenError("0a", &IdentifierToken{}))
	assert.Error(t, parseTokenError("*a", &IdentifierToken{}))
}

func TestDoubleStringToken(t *testing.T) {
	// double string
	assert.Nil(t, parseTokenError("abc\"", DoubleQuoteString()))
}

func TestMain(t *testing.T) {
	exp := New()
	_, err := exp.Parse("(123 + 5)|counter:1| ($a1 + 3)")
	assert.Error(t, err)
	assert.Nil(t, err)
}
