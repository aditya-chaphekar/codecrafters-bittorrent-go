package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
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

func decodeDictionary(bencodedString string) (interface{}, int, error) {
	dict := make(map[string]interface{})
	index := 1
	for index < len(bencodedString) {
		if bencodedString[index] == 'e' {
			return dict, index + 1, nil
		}
		key, keyLen, err := decodeString(bencodedString[index:])
		if err != nil {
			return nil, 0, err
		}
		value, valueLen, err := decodeBencode(bencodedString[index+keyLen:])
		if err != nil {
			return nil, 0, err
		}
		dict[key.(string)] = value
		index += keyLen + valueLen
	}
	return nil, 0, errors.New("invalid bencoded dictionary: missing 'e'")
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
	case 'd':
		return decodeDictionary(bencodedString)
	default:
		if unicode.IsDigit(rune(bencodedString[0])) {
			return decodeString(bencodedString)
		}
		return nil, 0, errors.New("invalid bencoded string")
	}
}
func extractMetadata(bencodedData map[string]interface{}) (string, int, map[string]interface{}, int, string, error) {
	announce, ok := bencodedData["announce"].(string)
	if !ok {
		return "", 0, nil, 0, "", errors.New("missing or invalid 'announce' field")
	}
	info, ok := bencodedData["info"].(map[string]interface{})
	if !ok {
		return "", 0, nil, 0, "", errors.New("missing or invalid 'info' field")
	}
	length, ok := info["length"].(int)
	if !ok {
		return "", 0, nil, 0, "", errors.New("missing or invalid 'length' field")
	}
	pieceLength, ok := info["piece length"].(int)
	if !ok {
		return "", 0, nil, 0, "", errors.New("missing or invalid 'piece length' field")
	}
	pieces, ok := info["pieces"].(string)
	if !ok {
		return "", 0, nil, 0, "", errors.New("missing or invalid 'pieces' field")
	}
	return announce, length, info, pieceLength, pieces, nil

}

func computeInfoHash(infoDict map[string]interface{}) (string, error) {
	// Sort the keys of the info dictionary
	sortedKeys := make([]string, 0, len(infoDict))
	for key := range infoDict {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)
	// Create a new map with the sorted keys
	sortedInfoDict := make(map[string]interface{})
	for _, key := range sortedKeys {
		sortedInfoDict[key] = infoDict[key]
	}
	// Bencode the sorted info dictionary
	bencodedInfo, _, err := encodeBencode(sortedInfoDict)
	if err != nil {
		return "", err
	}
	// Compute the SHA-1 hash of the bencoded info dictionary
	hash := sha1.New()
	hash.Write([]byte(bencodedInfo))
	infoHash := hash.Sum(nil)
	// Convert the hash to a hexadecimal string
	return hex.EncodeToString(infoHash), nil
}

// Function to bencode the data (similar to decode but for encoding)
func encodeBencode(data interface{}) (string, int, error) {
	switch v := data.(type) {
	case string:
		return fmt.Sprintf("%d:%s", len(v), v), len(v) + 2, nil
	case int:
		return fmt.Sprintf("i%de", v), len(fmt.Sprintf("i%de", v)), nil
	case []interface{}:
		result := "l"
		for _, item := range v {
			encoded, length, err := encodeBencode(item)
			if err != nil {
				return "", 0, err
			}
			result += encoded
			result = result + encoded[length:]
		}
		return result + "e", len(result) + 1, nil
	case map[string]interface{}:
		result := "d"
		// Sort the keys of the dictionary
		sortedKeys := make([]string, 0, len(v))
		for key := range v {
			sortedKeys = append(sortedKeys, key)
		}
		sort.Strings(sortedKeys)
		// Encode each key-value pair
		for _, key := range sortedKeys {
			encodedKey, _, err := encodeBencode(key)
			if err != nil {
				return "", 0, err
			}
			encodedValue, _, err := encodeBencode(v[key])
			if err != nil {
				return "", 0, err
			}
			result += encodedKey + encodedValue
		}
		return result + "e", len(result) + 1, nil
	}
	return "", 0, errors.New("unsupported type for bencoding")
}

func printPieceHashes(pieces string) {
	pieceCount := len(pieces) / 20
	for i := 0; i < pieceCount; i++ {
		hash := pieces[i*20 : (i+1)*20]
		fmt.Printf("%s\n", hex.EncodeToString([]byte(hash)))
	}
}

func main() {
	command := os.Args[1]

	switch command {
	case "decode":
		bencodedValue := os.Args[2]

		decoded, _, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
		break
	case "info":
		filePath := os.Args[2]
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		decoded, _, err := decodeBencode(string(fileData))
		if err != nil {
			fmt.Println("Error decoding file:", err)
			return
		}
		if dict, ok := decoded.(map[string]interface{}); ok {
			announce, length, infoDict, pieceLength, pieces, err := extractMetadata(dict)
			if err != nil {
				fmt.Println("Error extracting metadata:", err)
				return
			}
			infoHash, err := computeInfoHash(infoDict)
			if err != nil {
				fmt.Println("Error computing info hash:", err)
				return
			}
			// Print the tracker URL, file length, and info hash
			fmt.Printf("Tracker URL: %s\nLength: %d\nInfo Hash: %s\n", announce, length, infoHash)
			// Print the piece length and piece hashes
			fmt.Printf("Piece Length: %d\nPiece Hashes:\n", pieceLength)
			printPieceHashes(pieces)
		} else {
			fmt.Println("Decoded data is not a dictionary")
		}
		break
	default:
		fmt.Println("Unknown command specified")
	}
}
