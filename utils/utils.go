package utils

func IsUpperByte(bt byte) bool {
	return bt >= 65 && bt <= 90
}

func IsLowerByte(bt byte) bool {
	return bt >= 97 && bt <= 122
}

func IsEnLetterByte(bt byte) bool {
	return IsLowerByte(bt) || IsUpperByte(bt)
}
