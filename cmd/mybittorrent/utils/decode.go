package utils

import (
	"fmt"
	"strconv"
	"unicode"
)

func DecodeBencode(bencodedString string) (interface{}, int, error) {
	if bencodedString[0] == 'l' {
		return decodeList(bencodedString)
	} else if bencodedString[0] == 'i' {
		return decodeInt(bencodedString)
	} else if unicode.IsDigit(rune(bencodedString[0])) {
		return decodeString(bencodedString)
	} else {
		return "", 0, fmt.Errorf("Only strings are supported at the moment")
	}
}

func decodeString(bencodedString string) (string, int, error) {
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
		return "", 0, err
	}

	return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], len(lengthStr) + 1 + length, nil
}

func decodeInt(bencodedString string) (int, int, error) {
	var EOLIndex int

	for i := 0; i < len(bencodedString); i++ {
		if bencodedString[i] == 'e' {
			EOLIndex = i
			break
		}
	}

	res, err := strconv.Atoi(bencodedString[1:EOLIndex])

	return res, len(strconv.Itoa(res)) + 2, err
}

func decodeList(bencodedString string) ([]interface{}, int, error) {
	var list = make([]interface{}, 0)
	var i int
	globalOffset := 2

	for i = 1; i < len(bencodedString); {
		if bencodedString[i] == 'e' {
			break
		}

		decoded, offset, err := DecodeBencode(bencodedString[i:])
		i += offset
		globalOffset += offset
		if err != nil {
			return nil, 0, err
		}

		list = append(list, decoded)
	}

	return list, globalOffset, nil
}
