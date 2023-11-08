package torrent

import (
	"bytes"
	"crypto/sha1"
	"os"

	bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

func ParseTorrentFile(filename string) (*TorrentFileMeta, error) {
	file, err := os.Open(filename)
	if err != nil {
		return &TorrentFileMeta{}, err
	}
	defer file.Close()

	info := TorrentFile{}
	if err := bencode.Unmarshal(file, &info); err == nil {
		if infoHashBytes, err := torrentInfoHash(&info); err == nil {
			meta := &TorrentFileMeta{
				TorrentFileInfo: info,
				InfoHashBytes:   infoHashBytes,
			}
			return meta, nil
		} else {
			return &TorrentFileMeta{}, err
		}
	} else {
		return &TorrentFileMeta{}, err
	}
}

func torrentInfoHash(torrentFile *TorrentFile) ([]byte, error) {
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
