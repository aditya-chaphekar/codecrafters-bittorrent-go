package utils

import (
	"fmt"
	"strconv"
	"unicode"
)

func DecodeBencode(bencodedString string) (interface{}, error) {
	if bencodedString[0] == 'i' {
		return decodeInt(bencodedString)
	} else if unicode.IsDigit(rune(bencodedString[0])) {
		return decodeString(bencodedString)
	} else {
		return "", fmt.Errorf("Only strings are supported at the moment")
	}
}

func decodeString(bencodedString string) (string, error) {
	var firstColonIndex int

	for i := 0; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			firstColonIndex = i
			break
		}
	}

	lengthStr := bencodedString[:firstColonIndex]

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", err
	}

	return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
}

func decodeInt(bencodedString string) (int, error) {
	var EOLIndex int

	for i := 0; i < len(bencodedString); i++ {
		if bencodedString[i] == 'e' {
			EOLIndex = i
			break
		}
	}

	return strconv.Atoi(bencodedString[1:EOLIndex])
}
