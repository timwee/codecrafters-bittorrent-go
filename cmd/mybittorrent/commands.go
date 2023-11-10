package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"

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

	data, err := client.DownloadPiece(meta, peerAddress, pieceId)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return data, nil
}

func DownloadFileSubCommand(outputFilePath, torrentFileName string) {
	meta, err := torrent.ParseTorrentFile(torrentFileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := torrent.NewClient(meta, &torrent.Config{
		PeerId: PeerId,
		Port:   Port,
	})

	fmt.Println("Retrieve peers...")
	peersResponse, err := client.RequestPeers(meta)
	if err != nil {
		fmt.Println(err)
		return
	}
	peerAddress := peersResponse.Peers[0]
	// peerAddr := fmt.Sprintf("%s:%d", peer.IP, peer.Port)
	// cli := NewClient("00112233445566778899")
	data, err := client.DownloadFile(meta, peerAddress)
	if err != nil {
		fmt.Println(err)
		return
	}
	file, err := os.Create(outputFilePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.Write(data)
	fmt.Printf("Downloaded test.torrent to to %s\n", outputFilePath)
}
