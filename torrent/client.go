package torrent

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

const (
	MessageUnchoke     = 1
	MessageIntereseted = 2
	MessageBitfield    = 5
	MessageRequest     = 6
	MessagePiece       = 7
)
const (
	BlockSize int64 = 16 * 1024
)

type Config struct {
	PeerId string
	Port   int
}
type Client struct {
	Meta *TorrentFile
	*Config
	Conns map[string]net.Conn
}

func NewClient(meta *TorrentFile, config *Config) *Client {
	return &Client{
		Meta:   meta,
		Config: config,
		Conns:  make(map[string]net.Conn),
	}
}

type PeersRequest struct {
	InfoHash   string `json:"info_hash"`
	PeerId     string `json:"peer_id"`
	Port       int    `json:"port"`
	Uploaded   int    `json:"uploaded"`
	Downloaded int    `json:"downloaded"`
	Left       int    `json:"left"`
	Compact    int    `json:"compact"`
}
type PeersResponse struct {
	Interval int    `json:"interval"`
	RawPeers string `json:"peers"`
}

func ParsePeers(rawPeers string) []string {
	peers := make([]string, 0, 10)
	PEER_SIZE := 6
	buf := []byte(rawPeers)
	for index := 0; index+PEER_SIZE <= len(buf); index += PEER_SIZE {
		a := buf[index+0]
		b := buf[index+1]
		c := buf[index+2]
		d := buf[index+3]
		portBytes := []byte{0, 0, 0, 0, 0, 0, buf[index+4], buf[index+5]}
		port := binary.BigEndian.Uint64(portBytes)
		ip := fmt.Sprintf("%d.%d.%d.%d:%d", a, b, c, d, port)
		peers = append(peers, ip)
	}
	return peers
}

type PeersResult struct {
	Interval int
	Peers    []string
}

func (client *Client) RequestPeers(uploaded, downloaded, left, compact int) (*PeersResult, error) {
	infoHashBytes, err := hex.DecodeString(client.Meta.InfoHash)
	if err != nil {
		return nil, err
	}
	baseURL, _ := url.Parse(client.Meta.Announce)
	params := url.Values{}
	params.Add("info_hash", string(infoHashBytes))
	params.Add("peer_id", client.PeerId)
	params.Add("port", strconv.Itoa(client.Port))
	params.Add("uploaded", strconv.Itoa(uploaded))
	params.Add("downloaded", strconv.Itoa(downloaded))
	params.Add("left", strconv.Itoa(left))
	params.Add("compact", strconv.Itoa(compact))
	baseURL.RawQuery = params.Encode()
	response, err := http.Get(baseURL.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("bad response")
	}
	var peersResponse PeersResponse
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	bencodeBody := buf.String()
	data, err := DecodeBencode(bencodeBody)
	if err != nil {
		return nil, err
	}
	marshaledData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(marshaledData, &peersResponse)
	if err != nil {
		return nil, err
	}
	m := data.(map[string]string)
	p := m["peers"]
	peers := ParsePeers(p)
	return &PeersResult{
		Interval: peersResponse.Interval,
		Peers:    peers,
	}, nil
}
func (client *Client) Dial(peerAddress string) error {
	conn, err := net.Dial("tcp", peerAddress)
	if err != nil {
		return err
	}
	client.Conns[peerAddress] = conn
	return nil
}
func (client *Client) Close(peerAddress string) {
	conn, ok := client.Conns[peerAddress]
	if !ok {
		return
	}
	conn.Close()
}
func (client *Client) Handshake(peerAddress string) (string, error) {
	conn, ok := client.Conns[peerAddress]
	if !ok {
		return "", fmt.Errorf("no connection with peer address: %s", peerAddress)
	}
	infoHashBytes, err := hex.DecodeString(client.Meta.InfoHash)
	if err != nil {
		return "", err
	}
	// data := make([]byte, 0, 10)
	data := make([]byte, 0)
	data = append(data, []byte{19}...)
	data = append(data, []byte("BitTorrent protocol")...)
	data = append(data, []byte{0, 0, 0, 0, 0, 0, 0, 0}...)
	data = append(data, infoHashBytes...)
	data = append(data, []byte(client.Config.PeerId)...)
	_, err = conn.Write(data)
	if err != nil {
		return "", err

	}
	reply := make([]byte, 48+20)
	_, err = conn.Read(reply)
	if err != nil {
		return "", err
	}
	peerId := reply[48 : 48+20]

	return hex.EncodeToString(peerId), nil
}
func (client *Client) RecieveMessage(peerAddress string) (byte, []byte, error) {
	conn, ok := client.Conns[peerAddress]
	if !ok {
		return 0, nil, fmt.Errorf("no connection with peer address: %s", peerAddress)
	}
	lengthBytes := make([]byte, 4)
	if _, err := conn.Read(lengthBytes); err != nil {
		return 0, nil, err
	}
	length := binary.BigEndian.Uint32(lengthBytes)
	messageType := make([]byte, 1)
	if _, err := conn.Read(messageType); err != nil {
		return 0, nil, err
	}
	length--
	message := make([]byte, length)
	if _, err := io.ReadAtLeast(conn, message, int(length)); err != nil {
		return 0, nil, err
	}
	fmt.Println("recieve", messageType[0], len(message))
	return messageType[0], message, nil
}
func (client *Client) SendMessage(peerAddress string, messageType byte, message []byte) error {
	conn, ok := client.Conns[peerAddress]
	if !ok {
		return fmt.Errorf("no connection with peer address: %s", peerAddress)
	}
	length := uint32(len(message)) + 1
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)
	data := make([]byte, 0)
	data = append(data, lengthBytes...)
	data = append(data, messageType)
	data = append(data, message...)
	fmt.Println("sending", data)
	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}
func (client *Client) RecieveBitfield(peerAddress string) ([]byte, error) {
	messageType, message, err := client.RecieveMessage(peerAddress)
	if err != nil {
		return nil, err
	}
	if messageType != MessageBitfield {
		return nil, fmt.Errorf("wrong message type %d", messageType)
	}
	return message, nil
}
func (client *Client) RecieveUnchoke(peerAddress string) error {
	messageType, _, err := client.RecieveMessage(peerAddress)
	if err != nil {
		return err
	}
	if messageType != MessageUnchoke {
		return fmt.Errorf("wrong message type %d", messageType)
	}
	return nil
}
func (client *Client) SendInterested(peerAddress string) error {
	return client.SendMessage(peerAddress, MessageIntereseted, []byte{})
}
func (client *Client) SendRequest(peerAddress string, pieceIndex int, begin, length int) error {
	message := make([]byte, 12)
	binary.BigEndian.PutUint32(message[0:4], uint32(pieceIndex))
	binary.BigEndian.PutUint32(message[4:8], uint32(begin))
	binary.BigEndian.PutUint32(message[8:12], uint32(length))
	return client.SendMessage(peerAddress, MessageRequest, message)
}
func (client *Client) RecievePiece(peerAddress string) (uint32, uint32, []byte, error) {
	messageType, message, err := client.RecieveMessage(peerAddress)
	if err != nil {
		return 0, 0, nil, err
	}
	if messageType != MessagePiece {
		return 0, 0, nil, fmt.Errorf("wrong message type %d", messageType)
	}
	pieceIndex := binary.BigEndian.Uint32(message[0:4])
	begin := binary.BigEndian.Uint32(message[4:8])
	block := message[8:]
	return pieceIndex, begin, block, nil
}
func (client *Client) DownloadFile(peerAddress string, pieceIndex int) ([]byte, error) {
	pieceLength := client.Meta.Info.PieceLength
	length := client.Meta.Info.Length
	// last not whole piece
	if pieceIndex >= int(length/pieceLength) {
		pieceLength = length - (pieceLength * pieceIndex)
	}
	fmt.Printf("[RequestPiece] - Piece Length: %d - Length: %d - Piece Index: %d\n", pieceLength, length, pieceIndex)
	data := make([]byte, pieceLength)
	lastBlockSize := int64(pieceLength) % BlockSize
	piecesNum := (int64(pieceLength) - lastBlockSize) / BlockSize
	if lastBlockSize > 0 {
		piecesNum++
	}
	fmt.Printf("[requestPiece] - Piece Length: %d # of Pieces: %d\n", pieceLength, piecesNum)
	for i := int64(0); i < int64(pieceLength); i += int64(BlockSize) {
		length := BlockSize
		if i+int64(BlockSize) > int64(pieceLength) {
			fmt.Printf("reached last block, changing size to %d\n", lastBlockSize)
			length = int64(pieceLength) - i
			if length > BlockSize {
				length = BlockSize
			}
		}
		if err := client.SendRequest(peerAddress, pieceIndex, int(i), int(length)); err != nil {
			return nil, err
		}
		recievedPieceIndex, recievedBegin, recievedBlock, err := client.RecievePiece(peerAddress)
		if err != nil {
			return nil, err
		}
		if recievedPieceIndex != uint32(pieceIndex) {
			return nil, errors.New("mismatched piece index")
		}
		copy(data[recievedBegin:], recievedBlock)
	}
	return data, nil

}
