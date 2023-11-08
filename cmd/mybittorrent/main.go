package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/codecrafters-io/bittorrent-starter-go/torrent"
)

const (
	PeerId = "00112233445566778899"
	Port   = 6881
)

func printInfo(meta *torrent.TorrentFileMeta) {

	fmt.Printf("Tracker URL: %s", meta.TorrentFileInfo.Announce)
	fmt.Printf("Length: %d\n", meta.TorrentFileInfo.Info.Length)
	fmt.Printf("Info Hash: %x\n", meta.InfoHashBytes)
	fmt.Printf("Piece Length: %d\n", meta.TorrentFileInfo.Info.PieceLength)
	fmt.Println("Pieces Hashes:")
	for i := 0; i < len(meta.TorrentFileInfo.Info.Pieces); i += 20 {
		fmt.Printf("%x\n", meta.TorrentFileInfo.Info.Pieces[i:i+20])
	}
}

func printIPs(trackerResp torrent.TrackerResponse) {
	offset := 0
	for offset+6 <= len(trackerResp.Peers) {
		ip := net.IP(trackerResp.Peers[offset : offset+4])

		port := binary.BigEndian.Uint16([]byte(trackerResp.Peers[offset+4 : offset+6]))
		fmt.Printf("%s:%d\n", ip.String(), port)
		offset += 6
	}
}

func DecodeCommand(bencodedValue string) (string, error) {
	decoded, err := torrent.DecodeBencode(bencodedValue)
	if err != nil {
		return "", err
	}

	jsonOutput, err := json.Marshal(decoded)
	if err != nil {
		return "", err
	}
	return string(jsonOutput), nil
}

func InfoCommand(fileName string) error {
	torrentFileMeta, err := torrent.ParseTorrentFile(fileName)
	if err != nil {
		return err
	}

	printInfo(torrentFileMeta)
	return nil
}

func PeersCommand(fileName string) (torrent.TrackerResponse, error) {
	meta, err := torrent.ParseTorrentFile(fileName)
	if err != nil {
		return torrent.TrackerResponse{}, err
	}
	trackerResp, err := torrent.GetPeers(meta)
	if err != nil {
		return torrent.TrackerResponse{}, err
	}
	return trackerResp, nil
}

func HandshakeCommand(fileName string, peer string) (string, error) {
	meta, err := torrent.ParseTorrentFile(fileName)
	if err != nil {
		return "", err
	}

	peerId, err := torrent.SendHandshake(meta, peer)
	if err != nil {
		return "", err
	}
	return peerId, nil
}

func DownloadPieceSubcommand(torrentMetaFilePath string, pieceId int) ([]byte, error) {
	meta, err := torrent.ParseTorrentFile(torrentMetaFilePath)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	client := torrent.NewClient(meta, &torrent.Config{
		PeerId: PeerId,
		Port:   Port,
	})

	fmt.Println("Retrieve peers...")
	peersResponse, err := client.RequestPeers(meta)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	peerAddress := peersResponse.Peers[1]
	fmt.Printf("Connecting to %s...\n", peerAddress)
	if err := client.Dial(peerAddress); err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer client.Close(peerAddress)

	fmt.Println("Sending handshake...")
	_, err = client.Handshake(peerAddress, meta.InfoHashBytes)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println("Handshake is successful")
	fmt.Println("Waiting for 'bitfield'...")
	if _, err := client.RecieveBitfield(peerAddress); err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println("Recieved 'bitfield'...")
	fmt.Println("Sending 'interested'")
	if err := client.SendInterested(peerAddress); err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println("Sent 'interested'")
	fmt.Println("Wating for 'unchoke'...")
	if err := client.RecieveUnchoke(peerAddress); err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println("Recieved 'unchoke'")
	data, err := client.DownloadFile(peerAddress, pieceId)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return data, nil
}

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
