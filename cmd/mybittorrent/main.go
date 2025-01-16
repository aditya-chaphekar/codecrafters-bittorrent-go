package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Ensures gofmt doesn't remove the "os" encoding/json import (feel free to remove this!)
var _ = json.Marshal

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345

func decodeString(bencodedString string) (interface{}, int, error) {
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

func decodeInt(bencodedString string) (interface{}, int, error) {
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

func decodeList(bencodedString string) (interface{}, int, error) {
	retList := make([]interface{}, 0)
	index := 1
	for index < len(bencodedString) {
		if bencodedString[index] == 'e' {
			index += 1
			break
		}
		decoded, offset, err := decodeBencode(bencodedString[index:])
		if err != nil {
			return nil, 0, err
		}
		retList = append(retList, decoded)
		index += offset
	}
	return retList, index, nil
}

func decodeBencode(bencodedString string) (interface{}, int, error) {
	if len(bencodedString) == 0 {
		return nil, 0, errors.New("empty bencoded string")
	}
	switch bencodedString[0] {
	case 'i':
		return decodeInt(bencodedString)
	case 'l':
		return decodeList(bencodedString)
	// case 'd':
	// 	return decodeDictionary(bencodedString)
	default:
		if unicode.IsDigit(rune(bencodedString[0])) {
			return decodeString(bencodedString)
		}
		return nil, 0, errors.New("invalid bencoded string")
	}
}
func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, _, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
