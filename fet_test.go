package fet

import (
	"strings"
	"testing"

	"github.com/fefit/fet/types"
	"github.com/stretchr/testify/assert"
)

func TestCompile(t *testing.T) {
	curConf := &Config{
		Mode:           types.Smarty,
		TemplateDir:    "tests/smarty/templates",
		CompileDir:     "tests/smarty/views",
		LeftDelimiter:  "{%",
		RightDelimiter: "%}",
		UcaseField:     true,
		AutoRoot:       true,
	}
	fet, _ := New(curConf)
	assertOutputToBe := func(t *testing.T, tpl string, data interface{}, output string) {
		result, err := fet.Fetch(tpl, data)
		assert.Nil(t, err)
		assert.Equal(t, output, strings.TrimSpace(result))
	}
	// test for syntax
	helloFet := "hello fet!"
	helloFetChars := strings.Split(helloFet, "")
	t.Run("Test Smarty mode compile", func(t *testing.T) {
		assertOutputToBe(t, "hello.tpl", nil, helloFet)
		assertOutputToBe(t, "variable.tpl", nil, helloFet)
		assertOutputToBe(t, "strvar.tpl", nil, helloFet)
		assertOutputToBe(t, "keywordvar.tpl", nil, helloFet)
		assertOutputToBe(t, "foreach.tpl", map[string][]string{
			"Result": helloFetChars,
		}, helloFet)
		assertOutputToBe(t, "for.tpl", map[string][]string{
			"Result": helloFetChars,
		}, helloFet)
		assertOutputToBe(t, "slice.tpl", map[string][]string{
			"Result": helloFetChars,
		}, "hello")
		assertOutputToBe(t, "plus.tpl", nil, "5,5,5,5")
		assertOutputToBe(t, "minus.tpl", nil, "0,0,0,0")
		assertOutputToBe(t, "multiple.tpl", nil, "24,24,24,24")
		assertOutputToBe(t, "divide.tpl", nil, "8,8,8,8")
		assertOutputToBe(t, "minmax.tpl", nil, "1,2")
		assertOutputToBe(t, "mod.tpl", nil, "1.15")
	})
}
