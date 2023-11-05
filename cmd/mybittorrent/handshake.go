package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
)

func SendHandshake(torrentFile TorrentFile, infohash []byte, peer string) (string, error) {
	conn, err := net.Dial("tcp", peer)
	if err != nil {
		return "", fmt.Errorf("unable to dial peer %s: %w", peer, err)
	}
	defer conn.Close()
	var buffer bytes.Buffer
	var infoHashBytesArray [20]byte
	// infoHashBytesSlice, err := hex.DecodeString(string(infohash))
	// if err != nil {
	// 	return "", fmt.Errorf("unable to decode hex info hash bytes %s: %w", infohash, err)
	// }
	copy(infoHashBytesArray[:], infohash)
	peerHandshakeMessageRequest := &PeerHandshakeMessage{
		ProtocolLength: 19,
		Protocol:       [19]byte{'B', 'i', 't', 'T', 'o', 'r', 'r', 'e', 'n', 't', ' ', 'p', 'r', 'o', 't', 'o', 'c', 'o', 'l'},
		Reserved:       [8]byte{0, 0, 0, 0, 0, 0, 0, 0},
		InfoHash:       infoHashBytesArray,
		PeerId:         [20]byte{'0', '0', '1', '1', '2', '2', '3', '3', '4', '4', '5', '5', '6', '6', '7', '7', '8', '8', '9', '9'},
	}
	err = binary.Write(&buffer, binary.BigEndian, peerHandshakeMessageRequest)
	if err != nil {
		return "", fmt.Errorf("unable to write peer handshake message to buffer: %w", err)
	}
	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		return "", fmt.Errorf("unable to write handshake message to peer %s: %w", peer, err)
	}
	resp := make([]byte, 68)
	_, err = conn.Read(resp)
	if err != nil {
		return "", fmt.Errorf("unable to read handshake message from peer %s: %w", peer, err)
	}
	// fmt.Printf("Received %d bytes: %x\n", n, resp[:n])
	var peerHandshakeMessageResponse PeerHandshakeMessage
	err = binary.Read(bytes.NewReader(resp), binary.BigEndian, &peerHandshakeMessageResponse)
	if err != nil {
		return "", fmt.Errorf("unable to read handshake message from peer %s: %w", peer, err)
	}
	return hex.EncodeToString(peerHandshakeMessageResponse.PeerId[:]), nil
}
