package main

import (
	"fmt"
	"net"
	"os"
)

func ProcessInfo() {
	decoded := ReadFileAndDecode()
	if dict, ok := decoded.(map[string]interface{}); ok {
		announce, length, infoDict, pieceLength, pieces, err := ExtractMetadata(dict)
		if err != nil {
			fmt.Println("Error extracting metadata:", err)
			return
		}
		infoHash, err := ComputeInfoHash(infoDict)
		if err != nil {
			fmt.Println("Error computing info hash:", err)
			return
		}
		// Print the tracker URL, file length, and info hash
		fmt.Printf("Tracker URL: %s\nLength: %d\nInfo Hash: %s\n", announce, length, infoHash)
		// Print the piece length and piece hashes
		fmt.Printf("Piece Length: %d\nPiece Hashes:\n", pieceLength)
		PrintPieceHashes(pieces)
	} else {
		fmt.Println("Decoded data is not a dictionary")
	}
}

func ProcessPeersInfo() {
	decoded := ReadFileAndDecode()
	if dict, ok := decoded.(map[string]interface{}); ok {
		announce, _, infoDict, _, _, err := ExtractMetadata(dict)
		if err != nil {
			fmt.Println("Error extracting metadata:", err)
			return
		}
		infoHash, err := ComputeInfoHash(infoDict)
		if err != nil {
			fmt.Println("Error computing info hash:", err)
			return
		}
		peerID := GeneratePeerID()
		fileLength, ok := dict["info"].(map[string]interface{})["length"].(int)
		if !ok {
			fmt.Println("Error: missing file length in torrent metadata")
			return
		}
		stringPeerID := string(peerID[:])
		peers, err := QueryTracker(announce, ConvertToPercentEncoded(infoHash), stringPeerID, 6881, fileLength)
		if err != nil {
			fmt.Println("Error querying tracker:", err)
			return
		}
		for _, peer := range peers {
			fmt.Println(peer)
		}
	} else {
		fmt.Println("Decoded data is not a dictionary")
	}
}

func ProcessHandshake() {
	peerAddress := os.Args[3]
	decoded := ReadFileAndDecode()
	if dict, ok := decoded.(map[string]interface{}); ok {
		// Extract the metadata to get the announce and infoHash
		_, _, infoDict, _, _, err := ExtractMetadata(dict)
		if err != nil {
			fmt.Println("Error extracting metadata:", err)
			return
		}
		infoHash, err := ComputeInfoHash(infoDict)
		if err != nil {
			fmt.Println("Error computing info hash:", err)
			return
		}
		conn, err := net.Dial("tcp", peerAddress)
		if err != nil {
			fmt.Println("Error connecting to Peer:", err)
			return
		}
		defer conn.Close()
		// Perform handshake
		_, response, err := PerformHandshake(conn, infoHash)
		peerID := ExtractPeerId(response)
		if err != nil {
			fmt.Println("Error performing handshake:", err)
			return
		}
		// Print the received peer ID
		fmt.Printf("Peer ID: %s\n", peerID)
	} else {
		fmt.Println("Decoded data is not a dictionary")
	}
}

func ProcessDownloadPiece() {

}
