package main

import (
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
)

func DownloadPiece() {
	var filePath, outputPath, pieceIndex string

	if os.Args[2] == "-o" {
		outputPath = os.Args[3]
		filePath = os.Args[4]
		pieceIndex = os.Args[5]
	} else {
		outputPath = os.Args[2]
		filePath = os.Args[3]
		pieceIndex = os.Args[4]
	}
	// Read the torrent file
	decoded := ReadFileAndDecode(&filePath)

	if dict, ok := decoded.(map[string]interface{}); ok {
		// Extract the metadata to get the announce and infoHash
		announce, length, infoDict, pieceLength, _, err := ExtractMetadata(dict)
		if err != nil {
			fmt.Println("Error extracting metadata:", err)
			return
		}
		infoHash, err := ComputeInfoHash(infoDict)
		if err != nil {
			fmt.Println("Error computing info hash:", err)
			return
		}
		myPeerID := GeneratePeerID()
		fileLength, ok := dict["info"].(map[string]interface{})["length"].(int)
		if !ok {
			fmt.Println("Error: missing file length in torrent metadata")
			return
		}
		// Query the tracker for a list of peers
		peers, err := QueryTracker(announce, ConvertToPercentEncoded(infoHash), string(myPeerID[:]), 6881, fileLength)
		if err != nil {
			fmt.Println("Error querying tracker:", err)
			return
		}
		peerAddress := peers[0]
		conn, err := net.Dial("tcp", peerAddress)
		if err != nil {
			fmt.Println("Error connecting to Peer:", err)
			return
		}
		defer conn.Close() // Perform handshake
		_, _, err = PerformHandshake(conn, infoHash)
		if err != nil {
			fmt.Println("Error performing handshake:", err)
			return
		}
		// check bitfield
		if err := CheckRecievedMessage(conn, 5); err != nil {
			fmt.Println("Error checking bitfield:", err)
			return
		}
		// send interested message
		interestedMessage := CreatePeerMessage(2, []byte{})
		_, err = conn.Write(interestedMessage)
		if err != nil {
			fmt.Println("Error sending interested message:", err)
			return
		}
		// wait until unchoke message is recieved
		if err := CheckRecievedMessage(conn, 1); err != nil {
			fmt.Println("Error checking unchoke message:", err)
			return
		}
		pieceSize := pieceLength
		pieceCnt := int(math.Ceil(float64(length) / float64(pieceSize)))
		pieceIndexInt, err := strconv.Atoi(pieceIndex)
		if err != nil {
			fmt.Println("Error converting piece index to int:", err)
			return
		}
		if pieceIndexInt == pieceCnt-1 {
			pieceSize = length % pieceLength
		}
		blockSize := 16 * 1024
		blockCnt := int(math.Ceil(float64(pieceSize) / float64(blockSize)))
		var data []byte
		for i := 0; i < blockCnt; i++ {
			blockData, err := ReadBlock(conn, i, blockSize, pieceSize, pieceIndexInt)
			if err != nil {
				fmt.Println("Error reading block:", err)
				return
			}
			data = append(data, blockData...)
		}
		file, err := os.Create(outputPath)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		_, err = file.Write(data)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Piece downloaded to %s.\n", outputPath)

	} else {
		fmt.Println("Decoded data is not a dictionary")
	}
}

func Download() {
	var filePath, outputPath string

	if os.Args[2] == "-o" {
		outputPath = os.Args[3]
		filePath = os.Args[4]
	} else {
		outputPath = os.Args[2]
		filePath = os.Args[3]
	}
	// Read the torrent file
	decoded := ReadFileAndDecode(&filePath)

	if dict, ok := decoded.(map[string]interface{}); ok {
		// Extract the metadata to get the announce and infoHash
		announce, length, infoDict, pieceLength, _, err := ExtractMetadata(dict)
		if err != nil {
			fmt.Println("Error extracting metadata:", err)
			return
		}
		infoHash, err := ComputeInfoHash(infoDict)
		if err != nil {
			fmt.Println("Error computing info hash:", err)
			return
		}
		myPeerID := GeneratePeerID()
		fileLength, ok := dict["info"].(map[string]interface{})["length"].(int)
		if !ok {
			fmt.Println("Error: missing file length in torrent metadata")
			return
		}
		// Query the tracker for a list of peers
		peers, err := QueryTracker(announce, ConvertToPercentEncoded(infoHash), string(myPeerID[:]), 6881, fileLength)
		if err != nil {
			fmt.Println("Error querying tracker:", err)
			return
		}
		peerAddress := peers[0]
		conn, err := net.Dial("tcp", peerAddress)
		if err != nil {
			fmt.Println("Error connecting to Peer:", err)
			return
		}
		defer conn.Close() // Perform handshake
		_, _, err = PerformHandshake(conn, infoHash)
		if err != nil {
			fmt.Println("Error performing handshake:", err)
			return
		}
		// check bitfield
		if err := CheckRecievedMessage(conn, 5); err != nil {
			fmt.Println("Error checking bitfield:", err)
			return
		}
		// send interested message
		interestedMessage := CreatePeerMessage(2, []byte{})
		_, err = conn.Write(interestedMessage)
		if err != nil {
			fmt.Println("Error sending interested message:", err)
			return
		}
		// wait until unchoke message is recieved
		if err := CheckRecievedMessage(conn, 1); err != nil {
			fmt.Println("Error checking unchoke message:", err)
			return
		}
		pieceSize := pieceLength
		pieceCnt := int(math.Ceil(float64(length) / float64(pieceSize)))
		var data []byte
		for i := 0; i < pieceCnt; i++ {
			if i == pieceCnt-1 {
				pieceSize = length % pieceLength
			}
			blockSize := 16 * 1024
			blockCnt := int(math.Ceil(float64(pieceSize) / float64(blockSize)))
			var blockDataByte []byte
			for j := 0; j < blockCnt; j++ {
				blockData, err := ReadBlock(conn, j, blockSize, pieceSize, i)
				if err != nil {
					if err.Error() == "EOF" {
						break
					}
					fmt.Println("Error reading block:", err)
					return
				}
				blockDataByte = append(blockDataByte, blockData...)
			}
			data = append(data, blockDataByte...)
		}
		file, err := os.Create(outputPath)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		_, err = file.Write(data)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("File downloaded to %s.\n", outputPath)
	} else {
		fmt.Println("Decoded data is not a dictionary")
	}
}
