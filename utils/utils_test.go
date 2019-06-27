package utils

import (
	"testing"

	"github.com/fefit/fet/types"
	"github.com/stretchr/testify/assert"
)

func TestIsIdentifier(t *testing.T) {
	fet := types.Gofet
	smarty := types.Smarty
	assert.True(t, IsIdentifier("a", fet))
	assert.True(t, IsIdentifier("a1", fet))
	assert.True(t, IsIdentifier("_a", fet))
	assert.False(t, IsIdentifier("_", fet))
	assert.False(t, IsIdentifier("1", fet))
	assert.False(t, IsIdentifier("$a", fet))
	assert.False(t, IsIdentifier("a", smarty))
	assert.True(t, IsIdentifier("$1", smarty))
	assert.True(t, IsIdentifier("$_", smarty))
	assert.False(t, IsIdentifier("$$", smarty))
}
