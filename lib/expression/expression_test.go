package expression

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var exp = New()

func assertTokenList(t *testing.T, code string, types ...string) {
	tokens, err := exp.tokenize(code)
	assert.Nil(t, err, "tokenize failed")
	assert.Equal(t, len(tokens), len(types), "the tokens's count is not equal")
	for i, token := range tokens {
		actualType := typeOf(token)
		toBeType := types[i]
		assert.Equal(t, actualType, toBeType, "the type is not matched")
	}
}

func assertErrorTokenize(t *testing.T, code string) {
	_, err := exp.tokenize(code)
	assert.Equal(t, true, err != nil)
}

func TestTokenize(t *testing.T) {
	// single
	t.Run("Test simple tokens", func(t *testing.T) {
		// string
		assertTokenList(t, `"hello"`, "StringToken")
		assertTokenList(t, `"hello 'world'"`, "StringToken")
		assertTokenList(t, `"hello \"world\""`, "StringToken")
		assertErrorTokenize(t, `'hello'`)
		// number
		assertTokenList(t, "1", "NumberToken")
		assertTokenList(t, "-1", "NumberToken")
		assertTokenList(t, "+1", "NumberToken")
		assertTokenList(t, "1.0", "NumberToken")
		assertTokenList(t, "-1.0", "NumberToken")
		assertTokenList(t, "+1.0", "NumberToken")
		assertTokenList(t, "1e10", "NumberToken")
		assertTokenList(t, "0.1", "NumberToken")
		assertTokenList(t, "1e+10", "NumberToken")
		assertTokenList(t, "1e-10", "NumberToken")
		assertTokenList(t, "1.0e10", "NumberToken")
		assertTokenList(t, "0b1010", "NumberToken")
		assertTokenList(t, "0xfae0", "NumberToken")
		assertTokenList(t, "0o644", "NumberToken")
		// identifier
		assertTokenList(t, "id", "IdentifierToken")
		assertTokenList(t, "_id", "IdentifierToken")
		assertTokenList(t, "i1", "IdentifierToken")
		assertTokenList(t, "_1", "IdentifierToken")
		assertTokenList(t, "ID", "IdentifierToken")
		assertTokenList(t, "i_", "IdentifierToken")
		// brackets
		assertTokenList(t, "(e)", "LeftBracketToken", "IdentifierToken", "RightBracketToken")
		assertTokenList(t, "a[0]", "IdentifierToken", "LeftSquareBracketToken", "NumberToken", "RightSquareBracketToken")
		// operators
		assertTokenList(t, "1+1", "NumberToken", "OperatorToken", "NumberToken")
		assertTokenList(t, "1-1", "NumberToken", "OperatorToken", "NumberToken")
		assertTokenList(t, "-1-1", "NumberToken", "OperatorToken", "NumberToken")
		assertTokenList(t, "-1-+1", "NumberToken", "OperatorToken", "NumberToken")
		assertTokenList(t, "-1+-1", "NumberToken", "OperatorToken", "NumberToken")
		assertTokenList(t, "!1", "OperatorToken", "NumberToken")
		assertTokenList(t, "1 bitor 1", "NumberToken", "SpaceToken", "OperatorToken", "SpaceToken", "NumberToken")
		assertTokenList(t, "not 1", "OperatorToken", "SpaceToken", "NumberToken")
	})
	// wrong tokens
	t.Run("Test wrong simple tokens", func(t *testing.T) {
		// // wrong string
		assertErrorTokenize(t, `'hello'`)
		assertErrorTokenize(t, `"hello`)
		assertErrorTokenize(t, `hello"`)
		// wrong number
		assertErrorTokenize(t, "01")
		assertErrorTokenize(t, "1e")
		assertErrorTokenize(t, "1e0b1")
		assertErrorTokenize(t, "1e1.5")
		assertErrorTokenize(t, "1e1e1")
		assertErrorTokenize(t, "0.e1")
		// wrong identifier
		assertErrorTokenize(t, "_")
		assertErrorTokenize(t, "1a")
		assertErrorTokenize(t, "1_")
		assertErrorTokenize(t, ".a")
		// wrong operators
		assertErrorTokenize(t, ">=1")
		assertErrorTokenize(t, `1 >= "a"`)
		assertErrorTokenize(t, `1bitor 2`)
		assertErrorTokenize(t, `1 bitor2`)
	})
	// complex tokens
	t.Run("Test multiple tokens", func(t *testing.T) {
		assertTokenList(t, "$time|strtotime|date_format:\"Y-m-d\"", "IdentifierToken", "OperatorToken", "IdentifierToken", "OperatorToken", "IdentifierToken", "OperatorToken", "StringToken")
	})
}

func TestToAst(t *testing.T) {
	t.Run("Test to ast", func(t *testing.T) {
		_, err := exp.Parse("!!!!!$a.b != \"1\"")
		assert.Nil(t, err)
	})
}
