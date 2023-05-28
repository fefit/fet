package utils

import "unicode"

func IsUpperByte(bt byte) bool {
	return bt >= 65 && bt <= 90
}

func IsLowerByte(bt byte) bool {
	return bt >= 97 && bt <= 122
}

func IsEnLetterByte(bt byte) bool {
	return IsLowerByte(bt) || IsUpperByte(bt)
}

/**
 * same bytes
 */
func IsSameBytes(a []byte, b []byte) bool {
	total := len(a)
	if total == len(b) {
		for i := 0; i < total; i++ {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}
	return false
}

func IsSpaceByte(bt byte) bool {
	return unicode.IsSpace(rune(bt))
}
