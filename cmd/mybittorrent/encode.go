package main

import (
	"errors"
	"fmt"
	"sort"
)

// Function to bencode the data (similar to decode but for encoding)
func EncodeBencode(data interface{}) (string, int, error) {
	switch v := data.(type) {
	case string:
		return fmt.Sprintf("%d:%s", len(v), v), len(v) + 2, nil
	case int:
		return fmt.Sprintf("i%de", v), len(fmt.Sprintf("i%de", v)), nil
	case []interface{}:
		result := "l"
		for _, item := range v {
			encoded, length, err := EncodeBencode(item)
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
			encodedKey, _, err := EncodeBencode(key)
			if err != nil {
				return "", 0, err
			}
			encodedValue, _, err := EncodeBencode(v[key])
			if err != nil {
				return "", 0, err
			}
			result += encodedKey + encodedValue
		}
		return result + "e", len(result) + 1, nil
	}
	return "", 0, errors.New("unsupported type for bencoding")
}
