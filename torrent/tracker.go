package torrent

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"
	"github.com/mirstar13/go-bittorrent/peers"
)

type trackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func (tr *Torrent) buildTrackerUrl(peerId [20]byte, port uint16) (string, error) {
	base, err := url.Parse(tr.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(tr.InfoHash[:])},
		"peer_id":    []string{string(peerId[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{strconv.Itoa(tr.Length)},
		"compact":    []string{"1"},
	}
	base.RawQuery = params.Encode()

	return base.String(), nil
}

func (tr *Torrent) RequestPeers(peerId [20]byte, port uint16) ([]peers.Peer, error) {
	url, err := tr.buildTrackerUrl(peerId, port)
	if err != nil {
		return nil, err
	}

	cl := &http.Client{Timeout: 10 * time.Second}

	resp, err := cl.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	trckrResp := trackerResp{}
	err = bencode.Unmarshal(resp.Body, &trckrResp)
	if err != nil {
		return nil, err
	}

	resultPeers := peers.Unmarshal([]byte(trckrResp.Peers))

	return resultPeers, nil
}
