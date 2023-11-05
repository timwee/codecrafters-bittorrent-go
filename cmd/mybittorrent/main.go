package main

import (
	// Uncomment this line to pass the first stage
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"unicode"

	bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
func decodeBencode(bencodedString string) (interface{}, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		var firstColonIndex int

		for i := 0; i < len(bencodedString); i++ {
			if bencodedString[i] == ':' {
				firstColonIndex = i
				break
			}
		}

		lengthStr := bencodedString[:firstColonIndex]

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return "", err
		}

		return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
	} else if number, intErr := decodeBenEncdoedNumber(bencodedString); intErr == nil {
		return number, intErr
	} else if data, err := bencode.Decode(strings.NewReader(bencodedString)); err == nil {
		return data, err
		// tr
	} else {
		return "", fmt.Errorf("only strings are supported at the moment")
	}
}

func decodeBenEncdoedNumber(bencodedString string) (int, error) {
	strSize := len(bencodedString)
	if strSize < 3 {
		return -1, fmt.Errorf("not long enough for ben encoded int")
	}
	if bencodedString[0] != 'i' || bencodedString[strSize-1] != 'e' {
		return -1, fmt.Errorf("doesn't have i and e surrounding the string")
	}
	return strconv.Atoi(bencodedString[1 : strSize-1])

}

func printInfo(torrent TorrentFile, hash []byte) {

	fmt.Printf("Tracker URL: %s", torrent.Announce)
	fmt.Printf("Length: %d\n", torrent.Info.Length)
	fmt.Printf("Info Hash: %x\n", hash)
	fmt.Printf("Piece Length: %d\n", torrent.Info.PieceLength)
	fmt.Println("Pieces Hashes:")
	for i := 0; i < len(torrent.Info.Pieces); i += 20 {
		fmt.Printf("%x\n", torrent.Info.Pieces[i:i+20])
	}
}

func ParseTorrentFile(filename string) (TorrentFile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return TorrentFile{}, err
	}
	defer file.Close()

	info := TorrentFile{}
	if err := bencode.Unmarshal(file, &info); err == nil {
		return info, nil
	} else {
		return TorrentFile{}, err
	}
}

func torrentInfoHash(torrentFile TorrentFile) ([]byte, error) {
	var buf bytes.Buffer
	marshalErr := bencode.Marshal(&buf, torrentFile.Info)
	if marshalErr != nil {
		return nil, marshalErr
	}
	hasher := sha1.New()
	hasher.Write(buf.Bytes())
	// shaInfo := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	shaInfo := hasher.Sum(nil)
	return shaInfo, nil
}

func printIPs(trackerResp TrackerResponse) {
	offset := 0
	for offset+6 <= len(trackerResp.Peers) {
		ip := net.IP(trackerResp.Peers[offset : offset+4])

		port := binary.BigEndian.Uint16([]byte(trackerResp.Peers[offset+4 : offset+6]))
		fmt.Printf("%s:%d\n", ip.String(), port)
		offset += 6
	}
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		fileName := os.Args[2]
		torrentFile, err := ParseTorrentFile(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}
		hash, err := torrentInfoHash(torrentFile)
		if err != nil {
			fmt.Println(err)
			return
		}

		printInfo(torrentFile, hash)

	} else if command == "peers" {
		fileName := os.Args[2]
		torrentFile, err := ParseTorrentFile(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}
		hash, err := torrentInfoHash(torrentFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		trackerResp, err := GetPeers(torrentFile, hash)
		if err != nil {
			fmt.Println(err)
			return
		}
		printIPs(trackerResp)

	} else if command == "handshake" {

		fileName := os.Args[2]
		torrentFile, err := ParseTorrentFile(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}
		hash, err := torrentInfoHash(torrentFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		peerId, err := SendHandshake(torrentFile, hash, os.Args[3])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Peer ID: %s\n", peerId)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
