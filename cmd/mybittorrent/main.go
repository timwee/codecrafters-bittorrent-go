package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	command := os.Args[1]
	switch command {
	case "decode":
		output, err := DecodeCommand(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(output)
	case "info":
		err := InfoCommand(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
	case "peers":
		output, err := PeersCommand(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		printIPs(output)

	case "handshake":
		peerId, err := HandshakeCommand(os.Args[2], os.Args[3])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Peer ID: %s\n", peerId)

	case "download_piece":
		outputFilePath := os.Args[3]
		torrentMetaFilePath := os.Args[4]
		pieceIndex, err := strconv.Atoi(os.Args[5])
		if err != nil {
			panic(err)
		}
		output, err := DownloadPieceSubcommand(torrentMetaFilePath, pieceIndex)
		if err != nil {
			panic(err)
		}
		file, err := os.Create(outputFilePath)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		file.Write(output)

		fmt.Printf("Piece %d downloaded to %s\n", pieceIndex, outputFilePath)
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
