package torrent

import "strconv"

type TorrentFileInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

type TorrentFile struct {
	Announce string
	Info     TorrentFileInfo
	InfoHash string
}

type TorrentFileMeta struct {
	TorrentFileInfo TorrentFile
	InfoHashBytes   []byte
}

type TrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

type TrackerClientParams struct {
	InfoHash   string `url:"info_hash,omitempty"`
	PeerId     string `url:"peer_id,omitempty"`
	Port       string `url:"port,omitempty"`
	Uploaded   string `url:"uploaded,omitempty"`
	Downloaded string `url:"downloaded,omitempty"`
	Left       string `url:"left,omitempty"`
	Compact    string `url:"compact,omitempty"`
}

func DefaultTrackerClientParams(infoHash string, fileLength int) *TrackerClientParams {
	return &TrackerClientParams{
		InfoHash:   infoHash,
		PeerId:     "00112233445566778899",
		Port:       "6881",
		Uploaded:   "0",
		Downloaded: "0",
		Left:       strconv.Itoa(fileLength),
		Compact:    "1",
	}
}

type PeerHandshakeMessage struct {
	ProtocolLength uint8
	Protocol       [19]byte
	Reserved       [8]byte
	InfoHash       [20]byte
	PeerId         [20]byte
}
