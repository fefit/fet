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
	assert.Nil(t, parseTokenError("0", NewNumberToken()))
	assert.Nil(t, parseTokenError("9", NewNumberToken()))
	assert.Nil(t, parseTokenError("3", NewNumberToken()))
	assert.Nil(t, parseTokenError("520", NewNumberToken()))
	assert.Nil(t, parseTokenError("5_20", NewNumberToken()))
	assert.Nil(t, parseTokenError("52_0", NewNumberToken()))
	assert.Nil(t, parseTokenError("5_2_0", NewNumberToken()))
	assert.Error(t, parseTokenError("5__2_0", NewNumberToken()))
	assert.Error(t, parseTokenError("5_2_0_ ", NewNumberToken()))
	// float
	assert.Nil(t, parseTokenError("0.3", NewNumberToken()))
	assert.Nil(t, parseTokenError("9.275_2", NewNumberToken()))
	assert.Nil(t, parseTokenError("3.12", NewNumberToken()))
	assert.Nil(t, parseTokenError("520.5_3", NewNumberToken()))
	assert.Nil(t, parseTokenError("5_20.1_2_3", NewNumberToken()))
	assert.Nil(t, parseTokenError("52_0.87", NewNumberToken()))
	assert.Nil(t, parseTokenError("5_2_0.12_3", NewNumberToken()))
	assert.Error(t, parseTokenError("1_3._1", NewNumberToken()))
	assert.Error(t, parseTokenError("1_3.1__1 ", NewNumberToken()))
	assert.Error(t, parseTokenError("1_3.1_ ", NewNumberToken()))
	// exponent
	assert.Nil(t, parseTokenError("0.3e1", NewNumberToken()))
	assert.Nil(t, parseTokenError("9.275_2E-1", NewNumberToken()))
	assert.Nil(t, parseTokenError("3.1_2e+00_1", NewNumberToken()))
	assert.Nil(t, parseTokenError("12E10", NewNumberToken()))
	assert.Nil(t, parseTokenError("1_2E1_8", NewNumberToken()))
	assert.Error(t, parseTokenError("1_e1", NewNumberToken()))
	assert.Error(t, parseTokenError("1_3.e1", NewNumberToken()))
	assert.Error(t, parseTokenError("1_3.1_e1", NewNumberToken()))
	assert.Error(t, parseTokenError("1_3.e_1", NewNumberToken()))
	assert.Error(t, parseTokenError("1_3.e-_1", NewNumberToken()))
	assert.Error(t, parseTokenError("1_3.e+_1", NewNumberToken()))
	assert.Error(t, parseTokenError("1_3.e+1_ ", NewNumberToken()))
	// binary number
	assert.Nil(t, parseTokenError("0b1", NewNumberToken()))
	assert.Nil(t, parseTokenError("0b101", NewNumberToken()))
	assert.Nil(t, parseTokenError("0b1_01", NewNumberToken()))
	assert.Nil(t, parseTokenError("0b1_0_1", NewNumberToken()))
	assert.Nil(t, parseTokenError("0B1", NewNumberToken()))
	assert.Nil(t, parseTokenError("0B101", NewNumberToken()))
	// octal number
	assert.Nil(t, parseTokenError("01", NewNumberToken()))
	assert.Nil(t, parseTokenError("001", NewNumberToken()))
	assert.Nil(t, parseTokenError("0_01", NewNumberToken()))
	assert.Nil(t, parseTokenError("0_0_1", NewNumberToken()))
	assert.Nil(t, parseTokenError("0o1067", NewNumberToken()))
	assert.Nil(t, parseTokenError("0o1_06_7", NewNumberToken()))
	assert.Nil(t, parseTokenError("0o755", NewNumberToken()))
	assert.Nil(t, parseTokenError("0O362", NewNumberToken()))
	assert.Nil(t, parseTokenError("0O577", NewNumberToken()))
	// hex number
	assert.Nil(t, parseTokenError("0x19f", NewNumberToken()))
	assert.Nil(t, parseTokenError("0xff_f", NewNumberToken()))
	assert.Nil(t, parseTokenError("0x1a_F_c", NewNumberToken()))
	assert.Nil(t, parseTokenError("0x3_e_1", NewNumberToken()))
	assert.Nil(t, parseTokenError("0X1_9_F", NewNumberToken()))
	assert.Nil(t, parseTokenError("0X302f", NewNumberToken()))
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

func TestMain(t *testing.T) {
	exp := New()
	_, err := exp.Parse("++$a + b.f * (c - e) / (5 - e) + 3")
	assert.Error(t, err)
	assert.Nil(t, err)
}
