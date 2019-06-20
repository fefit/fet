package expression

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

var exp = New()

func TestTokenize(t *testing.T) {
	assertTokenList := func(code string, types ...string) {
		tokens, err := exp.tokenize(code)
		assert.Nil(t, err, "tokenize success")
		assert.Equal(t, len(tokens), len(types), "the tokens's count is not equal")
		for i, token := range tokens {
			actualType := typeOf(token)
			toBeType := types[i]
			log.Println(actualType, toBeType)
			assert.Equal(t, actualType, toBeType, "the type is not matched")
		}
	}
	assertErrorTokenize := func(code string) {
		_, err := exp.tokenize(code)
		assert.Equal(t, true, err != nil)
	}
	// string
	assertTokenList(`"hello"`, "StringToken")
	assertTokenList(`"hello 'world'"`, "StringToken")
	assertTokenList(`"hello \"world\""`, "StringToken")
	// number
	assertTokenList("1", "NumberToken")
	assertTokenList("-1", "NumberToken")
	assertTokenList("+1", "NumberToken")
	assertTokenList("1.0", "NumberToken")
	assertTokenList("-1.0", "NumberToken")
	assertTokenList("+1.0", "NumberToken")
	assertTokenList("1e10", "NumberToken")
	assertTokenList("0.1", "NumberToken")
	assertTokenList("1e+10", "NumberToken")
	assertTokenList("1e-10", "NumberToken")
	assertTokenList("1.0e10", "NumberToken")
	assertTokenList("0b1010", "NumberToken")
	assertTokenList("0xfae0", "NumberToken")
	assertTokenList("0o644", "NumberToken")
	// identifier
	assertTokenList("id", "IdentifierToken")
	assertTokenList("_id", "IdentifierToken")
	assertTokenList("i1", "IdentifierToken")
	assertTokenList("_1", "IdentifierToken")
	assertTokenList("ID", "IdentifierToken")
	assertTokenList("i_", "IdentifierToken")
	// brackets
	assertTokenList("(e)", "LeftBracketToken", "IdentifierToken", "RightBracketToken")
	assertTokenList("a[0]", "IdentifierToken", "LeftSquareBracketToken", "NumberToken", "RightSquareBracketToken")
	// operators
	assertTokenList("1+1", "NumberToken", "OperatorToken", "NumberToken")
	assertTokenList("1-1", "NumberToken", "OperatorToken", "NumberToken")
	assertTokenList("-1-1", "NumberToken", "OperatorToken", "NumberToken")
	assertTokenList("-1-+1", "NumberToken", "OperatorToken", "NumberToken")
	assertTokenList("-1+-1", "NumberToken", "OperatorToken", "NumberToken")
	assertTokenList("!1", "OperatorToken", "NumberToken")
	assertTokenList("1 bitor 1", "NumberToken", "SpaceToken", "OperatorToken", "SpaceToken", "NumberToken")
	assertTokenList("not 1", "OperatorToken", "SpaceToken", "NumberToken")
	// wrong string
	assertErrorTokenize(`'hello'`)
	assertErrorTokenize(`"hello`)
	assertErrorTokenize(`hello"`)
	// wrong number
	assertErrorTokenize("01")
	assertErrorTokenize("1e")
	assertErrorTokenize("1e1.5")
}
