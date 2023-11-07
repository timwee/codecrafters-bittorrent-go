package torrent

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/dghubble/sling"
	"github.com/jackpal/bencode-go"
)

func GetPeers(torrentFile TorrentFile, infoHash []byte) (TrackerResponse, error) {
	params := DefaultTrackerClientParams(string(infoHash), torrentFile.Info.Length)
	req, err := sling.New().Get(torrentFile.Announce).QueryStruct(params).Request()
	if err != nil {
		return TrackerResponse{}, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return TrackerResponse{}, err
	}
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		return TrackerResponse{}, err
	}
	// fmt.Printf("client: response body: %s\n", resBody)

	var trackerResp TrackerResponse
	if err := bencode.Unmarshal(bytes.NewReader(resBody), &trackerResp); err != nil {
		return TrackerResponse{}, err
	}
	return trackerResp, nil
}
