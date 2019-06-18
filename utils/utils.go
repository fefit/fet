package utils

import "unicode"

func IsIdentifier(name string) bool {
	if name == "" || name == "_" {
		return false
	}
	runes := []rune(name)
	for i, total := 0, len(runes); i < total; i++ {
		cur := runes[i]
		if unicode.IsLetter(cur) || cur == '_' {
			continue
		} else if unicode.IsDigit(cur) && i > 0 {
			continue
		}
		return false
	}
	return true
}
