package fet

import (
	"strings"
	"testing"

	"github.com/fefit/fet/types"
	"github.com/stretchr/testify/assert"
)

func TestMergeConfig(t *testing.T) {
	curConf := &Config{
		LeftDelimiter:  "{{",
		RightDelimiter: "}}",
		Mode:           types.Gofet,
		TemplateDir:    "test_template",
		CompileDir:     "test_compile",
	}
	fet, _ := New(curConf)
	conf := fet.Config
	assert.Equal(t, conf.LeftDelimiter, curConf.LeftDelimiter)
	assert.Equal(t, conf.RightDelimiter, curConf.RightDelimiter)
	assert.Equal(t, conf.Mode, curConf.Mode)
	assert.Equal(t, conf.TemplateDir, curConf.TemplateDir)
	assert.Equal(t, conf.CompileDir, curConf.CompileDir)
}

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
	//
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
	})
}
