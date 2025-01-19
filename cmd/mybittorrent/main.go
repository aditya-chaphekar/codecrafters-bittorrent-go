package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"strconv"
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
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		decoded, _, err := DecodeBencode(string(fileData))
		if err != nil {
			fmt.Println("Error decoding file:", err)
			return
		}
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
			// Select a peer to download the piece fromc
			peerAddress := peers[1]
			conn, err := net.Dial("tcp", peerAddress)
			if err != nil {
				fmt.Println("Error connecting to Peer:", err)
				return
			}
			defer conn.Close()
			// Convert the piece index to an integer
			// pieceIndexInt, err := strconv.Atoi(pieceIndex)
			// if err != nil {
			// 	fmt.Println("Error converting piece index to int:", err)
			// 	return
			// }
			// Perform handshake
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
				blockOffset := i * blockSize
				blockSize := blockSize
				if blockOffset+blockSize > pieceSize {
					blockSize = pieceSize - blockOffset
				}
				if blockSize == 0 {
					break
				}
				var reqbuf bytes.Buffer
				binary.Write(&reqbuf, binary.BigEndian, uint32(pieceIndexInt))
				binary.Write(&reqbuf, binary.BigEndian, uint32(blockOffset))
				binary.Write(&reqbuf, binary.BigEndian, uint32(blockSize))
				requestMessage := CreatePeerMessage(6, reqbuf.Bytes())

				_, err = conn.Write(requestMessage)
				if err != nil {
					fmt.Println("Error sending request message:", err)
					return
				}
				// Read the payload consist of index, begin, block
				buf := make([]byte, 4)
				_, err = conn.Read(buf)
				if err != nil {
					fmt.Println(err)
					return
				}

				payloadBuf := make([]byte, binary.BigEndian.Uint32(buf))
				_, err = io.ReadFull(conn, payloadBuf)
				if err != nil {
					fmt.Println(err)
					return
				}
				if int(payloadBuf[0]) != 7 {
					fmt.Println("Unexpected message ID")
					return
				}
				data = append(data, payloadBuf[9:]...)

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
	case "download":
	default:
		fmt.Println("Unknown command specified")
	}
}
