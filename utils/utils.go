package utils

import (
	"unicode"

	"github.com/fefit/fet/types"
)

// IsIdentifier judge
func IsIdentifier(name string, mode types.Mode) bool {
	if name == "" || name == "_" || name == "$" {
		return false
	}
	runes := []rune(name)
	isDollar := false
	for i, total := 0, len(runes); i < total; i++ {
		cur := runes[i]
		startWithDollar := cur == '$' && i == 0
		if unicode.IsLetter(cur) || cur == '_' || startWithDollar {
			isDollar = startWithDollar || isDollar
			continue
		} else if unicode.IsDigit(cur) && i > 0 {
			continue
		}
		return false
	}
	if mode == types.AnyMode {
		return true
	}
	if mode&types.Smarty == types.Smarty {
		return isDollar
	}
	return !isDollar
}
