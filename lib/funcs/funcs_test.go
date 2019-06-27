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
	assert.True(t, empty(nil))
	assert.True(t, empty(""))
	assert.True(t, empty("0"))
	assert.False(t, empty(" "))
	assert.True(t, empty(false))
	assert.False(t, empty(true))
	assert.True(t, empty([]string{}))
	assert.False(t, empty([2]string{}))
	assert.True(t, empty(map[string]string{}))
	assert.False(t, empty(p, "Name"))
	assert.True(t, empty(p, "Score", "Min"))
	assert.True(t, empty(p, "Nickname"))
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, truncate("hello fet!", 10), "hello fet!")
	assert.Equal(t, truncate("hello fet!", 5), "hello...")
	assert.Equal(t, truncate("你hello好，fet", 5), "你hell...")
	assert.Equal(t, truncate("你好", 5), "你好")
}

func TestConcat(t *testing.T) {
	assert.Equal(t, concat("hello", " ", "fet!"), "hello fet!")
	assert.Equal(t, concat("你好", "fet!"), "你好fet!")
}

func TestCount(t *testing.T) {
	assert.Equal(t, count("你好fet!"), 6)
	assert.Equal(t, count([2]int{}), 2)
	assert.Equal(t, count([]int{1, 2, 3}), 3)
	assert.Equal(t, count(map[int]int{0: 1, 1: 2, 2: 3}), 3)
}

func TestMaths(t *testing.T) {
	assert.Equal(t, floor(1.5), 1.0)
	assert.Equal(t, ceil(1.5), 2.0)
}

func TestNumberformat(t *testing.T) {
	assert.Equal(t, numberFormat(10000), "10,000")
	assert.Equal(t, numberFormat(100), "100")
	assert.Equal(t, numberFormat(1000000), "1,000,000")
	assert.Equal(t, numberFormat(1000000, 1), "1,000,000.0")
	assert.Equal(t, numberFormat(1000000, 1, "@"), "1,000,000@0")
	assert.Equal(t, numberFormat(1000000, 1, "@", "z"), "1z000z000@0")
}
