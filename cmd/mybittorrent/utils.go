package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"math/rand"
)

func PrintPieceHashes(pieces string) {
	pieceCount := len(pieces) / 20
	for i := 0; i < pieceCount; i++ {
		hash := pieces[i*20 : (i+1)*20]
		fmt.Printf("%s\n", hex.EncodeToString([]byte(hash)))
	}
}

func ParsePeers(peers string) []string {
	var peerList []string
	for i := 0; i < len(peers); i += 6 {
		ip := fmt.Sprintf("%d.%d.%d.%d", peers[i], peers[i+1], peers[i+2], peers[i+3])
		port := int(peers[i+4])<<8 | int(peers[i+5])
		peerList = append(peerList, fmt.Sprintf("%s:%d", ip, port))
	}
	return peerList
}

func ConvertToPercentEncoded(input string) string {
	// Decode the hex string into a byte slice
	data, err := hex.DecodeString(input)
	if err != nil {
		panic(err)
	}
	// Convert each byte to a percent-encoded string
	var builder strings.Builder
	for _, b := range data {
		builder.WriteString(fmt.Sprintf("%%%02x", b))
	}
	return builder.String()
}

// Generate a random Peer ID (20 bytes)
func GeneratePeerID() [20]byte {
	var peerID [20]byte
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 20; i++ {
		peerID[i] = byte(rand.Intn(256))
	}
	return peerID
}

func ExtractPeerId(response []byte) string {
	return hex.EncodeToString(response[48:68])
}

func CheckRecievedMessage(conn net.Conn, expectedMessageID int) error {
	buf := make([]byte, 4)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	payloadBuf := make([]byte, binary.BigEndian.Uint32(buf))
	_, err = conn.Read(payloadBuf)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if int(payloadBuf[0]) != expectedMessageID {
		return errors.New("unexpected message ID")
	}
	return nil
}

func CreatePeerMessage(messageID byte, payload []byte) []byte {
	messageLength := make([]byte, 4)
	binary.BigEndian.PutUint32(messageLength, uint32(len(payload)+1))
	return append(append(messageLength, messageID), payload...)
}

func ReadTorrentFile(path string) []byte {
	fileData, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return fileData
}

func ReadFileAndDecode() interface{} {
	fileData := ReadTorrentFile(os.Args[2])
	decoded, _, err := DecodeBencode(string(fileData))
	if err != nil {
		fmt.Println("Error decoding file:", err)
		return nil
	}
	return decoded
}
