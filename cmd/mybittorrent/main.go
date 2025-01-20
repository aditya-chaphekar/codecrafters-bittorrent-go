package main

import (
	"encoding/json"
	"fmt"
	"os"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

func main() {
	command := os.Args[1]

	switch command {
	case "decode":
		bencodedValue := os.Args[2]
		decoded, _, err := DecodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}
		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	case "info":
		ProcessInfo()
	case "peers":
		ProcessPeersInfo()
	case "handshake":
		ProcessHandshake()
	case "download_piece":
		DownloadPiece()
	case "download":
		Download()
	default:
		fmt.Println("Unknown command specified")
	}
}
