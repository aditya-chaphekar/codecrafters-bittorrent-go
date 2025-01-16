package main

import (
	"encoding/json"
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

	return bencodedString[firstColonIdx+1 : firstColonIdx+1+length], len(lengthStr) + 1 + length, nil

}

func decodeInt(bencodedString string) (interface{}, int, error) {
	var EIdx int
	for i := 0; i < len(bencodedString); i++ {
		if bencodedString[i] == 'e' {
			EIdx = i
			break
		}
	}
	retInt, err := strconv.Atoi(bencodedString[1:EIdx])
	if err != nil {
		fmt.Println("Error Converting to Int", err)
	}

	return retInt, len(bencodedString[1:EIdx]), nil
}

func decodeBencode(bencodedString string) (interface{}, error) {
	i := 0
	strLen := len(bencodedString)
	var ret interface{}
	for i = 0; i < strLen; i++ {
		switch bencodedString[i] {
		case 'i':
			decodedInt, skip, err := decodeInt(bencodedString[i:])
			i = skip
			if err != nil {
				fmt.Println("Error Decoding Int", err)
			}
			ret = decodedInt
			break
		default:
			if unicode.IsDigit(rune(bencodedString[i])) {
				decodedString, _, err := decodeString(bencodedString)
				i = strLen
				if err != nil {
					fmt.Println("Error Decoding String", err)
				}
				ret = decodedString

			}
		}
	}
	return ret, nil
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
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
