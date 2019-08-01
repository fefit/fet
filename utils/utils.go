package utils

import (
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
		if IsEnLetter(cur) || cur == '_' || startWithDollar {
			isDollar = startWithDollar || isDollar
			continue
		} else if IsArabicNumber(cur) && i > 0 {
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

// IsEnLetter Judge if is english alphabet
func IsEnLetter(s rune) bool {
	return (s >= 97 && s <= 122) || (s >= 65 && s <= 90)
}

// IsArabicNumber Judge if is Arabic numbers
func IsArabicNumber(s rune) bool {
	return s >= 48 && s <= 57
}
