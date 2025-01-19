package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode"
)

func DecodeString(bencodedString string) (interface{}, int, error) {
	var firstColonIdx int
	for i := 0; i < len(bencodedString); i++ {
		if bencodedString[i] == ':' {
			firstColonIdx = i
			break
		}
	}

	lengthStr := bencodedString[:firstColonIdx]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", 0, err
	}
	start := firstColonIdx + 1
	end := start + length
	return bencodedString[start:end], end, nil
}

func DecodeInt(bencodedString string) (interface{}, int, error) {
	var EIdx int
	for i := 0; i < len(bencodedString); i++ {
		if bencodedString[i] == 'e' {
			EIdx = i
			break
		}
	}
	numberStr := bencodedString[1:EIdx]
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return nil, 0, err
	}
	return number, EIdx + 1, nil
}

func DecodeList(bencodedString string) (interface{}, int, error) {
	retList := make([]interface{}, 0)
	index := 1
	for index < len(bencodedString) {
		if bencodedString[index] == 'e' {
			index += 1
			break
		}
		decoded, offset, err := DecodeBencode(bencodedString[index:])
		if err != nil {
			return nil, 0, err
		}
		retList = append(retList, decoded)
		index += offset
	}
	return retList, index, nil
}

func DecodeDictionary(bencodedString string) (interface{}, int, error) {
	dict := make(map[string]interface{})
	index := 1
	for index < len(bencodedString) {
		if bencodedString[index] == 'e' {
			return dict, index + 1, nil
		}
		key, keyLen, err := DecodeString(bencodedString[index:])
		if err != nil {
			return nil, 0, err
		}
		value, valueLen, err := DecodeBencode(bencodedString[index+keyLen:])
		if err != nil {
			return nil, 0, err
		}
		dict[key.(string)] = value
		index += keyLen + valueLen
	}
	return nil, 0, errors.New("invalid bencoded dictionary: missing 'e'")
}

func DecodeBencode(bencodedString string) (interface{}, int, error) {
	if len(bencodedString) == 0 {
		return nil, 0, errors.New("empty bencoded string")
	}
	switch bencodedString[0] {
	case 'i':
		return DecodeInt(bencodedString)
	case 'l':
		return DecodeList(bencodedString)
	case 'd':
		return DecodeDictionary(bencodedString)
	default:
		if unicode.IsDigit(rune(bencodedString[0])) {
			return DecodeString(bencodedString)
		}
		return nil, 0, errors.New("invalid bencoded string")
	}
}

func DecodeBencodeResponse(body io.Reader) (map[string]interface{}, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	decoded, _, err := DecodeBencode(string(data))
	if err != nil {
		return nil, err
	}
	if dict, ok := decoded.(map[string]interface{}); ok {
		return dict, nil
	}
	return nil, fmt.Errorf("decoded response is not a dictionary")
}
