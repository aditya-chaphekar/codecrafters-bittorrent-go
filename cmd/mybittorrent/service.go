package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
)

func ExtractMetadata(bencodedData map[string]interface{}) (string, int, map[string]interface{}, int, string, error) {
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

func ComputeInfoHash(infoDict map[string]interface{}) (string, error) {
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
	bencodedInfo, _, err := EncodeBencode(sortedInfoDict)
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

func QueryTracker(trackerURL string, infoHash string, peerId string, port int, fileLength int) ([]string, error) {
	// Construct query parameters
	queryParams := url.Values{}
	//queryParams.Set("info_hash", infoHash) // URL-encode the raw info_hash bytes (this is correct)
	queryParams.Set("peer_id", peerId)
	queryParams.Set("port", strconv.Itoa(port))
	queryParams.Set("uploaded", "0")
	queryParams.Set("downloaded", "0")
	queryParams.Set("left", strconv.Itoa(fileLength))
	queryParams.Set("compact", "1")
	// Construct the full URL with query parameters
	fullURL := fmt.Sprintf("%s?%s&info_hash=%s", trackerURL, queryParams.Encode(), infoHash)
	// Send GET request to the tracker
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query tracker: %v", err)
	}
	defer resp.Body.Close()
	// Debugging: print the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tracker response: %v", err)
	}
	// Decode the bencoded response
	decodedResponse, err := DecodeBencodeResponse(bytes.NewReader(body)) // Pass body as Reader here
	if err != nil {
		return nil, fmt.Errorf("failed to decode tracker response: %v", err)
	}
	// Check if failure reason exists
	if failureReason, ok := decodedResponse["failure reason"].(string); ok {
		return nil, fmt.Errorf("tracker returned failure reason: %s", failureReason)
	}
	// Extract peer list from the response
	peers, ok := decodedResponse["peers"].(string)
	if !ok {
		return nil, fmt.Errorf("peers not found in tracker response")
	}
	// Parse the peer list (compact format) and return peer addresses
	return ParsePeers(peers), nil
}

// Perform the handshake with the peer
func PerformHandshake(conn net.Conn, infoHash string) (net.Conn, []byte, error) {
	// Prepare the handshake message
	protocol := "BitTorrent protocol"
	reserved := make([]byte, 8)                    // 8 reserved bytes set to zero
	peerID := GeneratePeerID()                     // Generate a random 20-byte peer ID
	infoHashBytes, _ := hex.DecodeString(infoHash) // Convert the hex infoHash to bytes
	// Construct the handshake message
	message := append([]byte{19}, []byte(protocol)...)
	message = append(message, reserved...)
	message = append(message, infoHashBytes...)
	message = append(message, peerID[:]...)
	// Send the handshake message
	_, err := conn.Write(message)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send handshake message: %v", err)
	}
	// Read the response (it should be the same format as the handshake)
	response := make([]byte, 68) // Expected size of a handshake response
	_, err = conn.Read(response)
	if err != nil {
		return nil, response, fmt.Errorf("failed to read handshake response: %v", err)
	}
	return conn, response, nil
}
