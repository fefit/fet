package funcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndex(t *testing.T) {
	p := struct {
		Name  string
		Score struct {
			Min int
			Max int
		}
	}{
		"hello fet!", struct {
			Min int
			Max int
		}{10, 100},
	}
	assert.Equal(t, index(p, "Name"), "hello fet!")
	assert.Equal(t, index(p, "Score", "Min"), 10)
	assert.Equal(t, index(&p, "Score", "Max"), 100)
	assert.Empty(t, index(p, "Score", 1))
	m := map[string]interface{}{
		"Name": "hello fet!",
		"Score": map[string]int{
			"Min": 10,
			"Max": 100,
		},
	}
	assert.Equal(t, index(m, "Name"), "hello fet!")
	assert.Equal(t, index(m, "Score", "Min"), 10)
	assert.Equal(t, index(&m, "Score", "Max"), 100)
	assert.Empty(t, index(m, "Score", 1))
	m2 := map[int8]interface{}{
		0: "hello fet!",
		1: &map[string]int{
			"Min": 10,
			"Max": 100,
		},
	}
	assert.Equal(t, index(m2, 0), "hello fet!")
	assert.Equal(t, index(m2, 1, "Min"), 10)
	assert.Equal(t, index(&m2, 1, "Max"), 100)
	assert.Empty(t, index(m2, "Name"))
	s := []interface{}{
		"hello fet!",
		map[string]int{
			"Min": 10,
			"Max": 100,
		},
	}
	assert.Equal(t, index(s, 0), "hello fet!")
	assert.Equal(t, index(s, 1, "Min"), 10)
	assert.Equal(t, index(&s, 1, "Max"), 100)
	assert.Empty(t, index(s, "Name"))
	a := [2]interface{}{
		"hello fet!",
		map[string]int{
			"Min": 10,
			"Max": 100,
		},
	}
	assert.Equal(t, index(a, 0), "hello fet!")
	assert.Equal(t, index(a, 1, "Min"), 10)
	assert.Equal(t, index(&a, 1, "Max"), 100)
	assert.Empty(t, index(a, "Name"))
}

func TestEmpty(t *testing.T) {
	p := struct {
		Name     string
		Nickname string
		Score    struct {
			Min int
			Max int
		}
	}{
		"hello fet!", "", struct {
			Min int
			Max int
		}{0, 100},
	}
	assert.Equal(t, true, empty(nil))
	assert.Equal(t, true, empty(""))
	assert.Equal(t, true, empty("0"))
	assert.Equal(t, false, empty(" "))
	assert.Equal(t, true, empty(false))
	assert.Equal(t, false, empty(true))
	assert.Equal(t, true, empty([]string{}))
	assert.Equal(t, false, empty([2]string{}))
	assert.Equal(t, true, empty(map[string]string{}))
	assert.Equal(t, false, empty(p, "Name"))
	assert.Equal(t, true, empty(p, "Score", "Min"))
	assert.Equal(t, true, empty(p, "Nickname"))
}
