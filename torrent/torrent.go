package torrent

import (
	"bytes"
	"crypto/sha1"
	"os"

	"github.com/jackpal/bencode-go"
	"github.com/mirstar13/go-bittorrent/p2p"
)

var (
	PeerId = [20]byte([]byte("glwe24pj7h19vw79cc2o"))
	Port   = uint16(6881)
)

type Torrent struct {
	Announce    string
	Name        string
	Length      int
	InfoHash    [20]byte
	PieceLength int
	PieceHashes [][20]byte
}

type TorrentInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
}

type TorrentFile struct {
	Announce  string      `bencode:"announce"`
	CreatedBy string      `bencode:"created by"`
	Info      TorrentInfo `bencode:"info"`
}

func (tr *TorrentFile) toTorrent() Torrent {
	infoHash := tr.Info.hash()
	pieceHashes := tr.Info.pieceHashes()

	return Torrent{
		Announce:    tr.Announce,
		Name:        tr.Info.Name,
		Length:      tr.Info.Length,
		InfoHash:    infoHash,
		PieceLength: tr.Info.PieceLength,
		PieceHashes: pieceHashes,
	}
}

func (info *TorrentInfo) hash() [20]byte {
	var buf bytes.Buffer
	bencode.Marshal(&buf, *info)
	h := sha1.Sum(buf.Bytes())
	return h
}

func (info *TorrentInfo) pieceHashes() [][20]byte {
	hashLen := 20
	buf := []byte(info.Pieces)

	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}

	return hashes
}

func (t *Torrent) DownloadFile(path string) error {
	peers, err := t.RequestPeers(PeerId, Port)
	if err != nil {
		return err
	}

	torrent := p2p.Torrent{
		PeerId:      PeerId,
		Peers:       peers,
		Name:        t.Name,
		InfoHash:    t.InfoHash,
		PieceLength: t.PieceLength,
		Length:      t.Length,
		PieceHashes: t.PieceHashes,
	}

	buf, err := torrent.DownloadFile()
	if err != nil {
		return err
	}

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = outFile.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

func (t *Torrent) DownloadPiece(index int, path string) error {
	peers, err := t.RequestPeers(PeerId, Port)
	if err != nil {
		return err
	}

	torrent := p2p.Torrent{
		PeerId:      PeerId,
		Peers:       peers,
		Name:        t.Name,
		InfoHash:    t.InfoHash,
		PieceLength: t.PieceLength,
		Length:      t.Length,
		PieceHashes: t.PieceHashes,
	}

	buf, err := torrent.DownloadPiece(index)
	if err != nil {
		return err
	}

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = outFile.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

func Open(fileName string) (Torrent, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return Torrent{}, err
	}
	defer file.Close()

	var torrentFile TorrentFile
	err = bencode.Unmarshal(file, &torrentFile)
	if err != nil {
		return Torrent{}, err
	}

	return torrentFile.toTorrent(), nil
}
