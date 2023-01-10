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
		Ignores:        []string{"tests/smarty/templates/inc"},
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
		assertOutputToBe(t, "strvar.tpl", nil, helloFet+helloFet)
		assertOutputToBe(t, "keywordvar.tpl", nil, helloFet)
		assertOutputToBe(t, "concat.tpl", nil, helloFet)
		// rewrite foreach
		assertOutputToBe(t, "foreach.tpl", map[string][]string{
			"Result": helloFetChars,
		}, helloFet)
		assertOutputToBe(t, "for.tpl", map[string][]string{
			"Result": helloFetChars,
		}, helloFet)
		assertOutputToBe(t, "slice.tpl", map[string][]string{
			"Result": helloFetChars,
		}, "hello")
		// maths
		assertOutputToBe(t, "plus.tpl", nil, "5,5,5,5")
		assertOutputToBe(t, "minus.tpl", nil, "0,0,0,0")
		assertOutputToBe(t, "multiple.tpl", nil, "24,24,24,24")
		assertOutputToBe(t, "divide.tpl", nil, "8,8,8,8")
		assertOutputToBe(t, "minmax.tpl", nil, "1,2")
		assertOutputToBe(t, "mod.tpl", nil, "1.15")
		assertOutputToBe(t, "power.tpl", nil, "1024")
		// pipe
		assertOutputToBe(t, "pipe.tpl", nil, "2021-09-05 18:07:06")
		// comment
		assertOutputToBe(t, "comment.tpl", nil, "")
		// include
		assertOutputToBe(t, "include.tpl", map[string]string{
			"Header": "hello",
			"Footer": "fet",
		}, "header:hello;footer:fet")
		assertOutputToBe(t, "include_props.tpl", nil, "header:hello;footer:fet")
		// extends
		assertOutputToBe(t, "extends.tpl", nil, "(header)\n(content)\n(footer)")
		assertOutputToBe(t, "extends_override.tpl", nil, "(override:header)\n(content)\n(footer)")
		// if condition
		assertOutputToBe(t, "condition.tpl", map[string]int{
			"Number": 1,
		}, "bigger")
		assertOutputToBe(t, "condition.tpl", map[string]int{
			"Number": 0,
		}, "equal")
		assertOutputToBe(t, "condition.tpl", map[string]int{
			"Number": -1,
		}, "smaller")
	})
}

func TestImportsDepends(t *testing.T) {
	var initImports = func() Imports {
		return Imports{
			Nodes: map[string]*ImportNode{},
		}
	}
	// A -> A depend itself
	imports := initImports()
	assert.True(t, len(imports.Add("A", "A")) > 0)
	// A -> B   B -> C  C -> A
	imports = initImports()
	assert.True(t, len(imports.Add("A", "B")) == 0)
	assert.True(t, len(imports.Add("B", "C")) == 0)
	assert.True(t, len(imports.Add("C", "A")) > 0)
	// A -> B  C -> A  B -> C
	imports = initImports()
	assert.True(t, len(imports.Add("A", "B")) == 0)
	assert.True(t, len(imports.Add("C", "A")) == 0)
	assert.True(t, len(imports.Add("B", "C")) > 0)
	// C -> D A -> B  D -> A  B -> C
	imports = initImports()
	assert.True(t, len(imports.Add("C", "D")) == 0)
	assert.True(t, len(imports.Add("A", "B")) == 0)
	assert.True(t, len(imports.Add("D", "A")) == 0)
	assert.True(t, len(imports.Add("B", "C")) > 0)
	// A -> B  A -> C  C -> D  D -> A
	imports = initImports()
	assert.True(t, len(imports.Add("A", "B")) == 0)
	assert.True(t, len(imports.Add("A", "C")) == 0)
	assert.True(t, len(imports.Add("D", "A")) == 0)
	assert.True(t, len(imports.Add("C", "D")) > 0)

}
