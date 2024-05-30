package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/mirstar13/go-bittorrent/client"
	"github.com/mirstar13/go-bittorrent/messages"
	"github.com/mirstar13/go-bittorrent/peers"
)

const MaxBlockSize = 16384

type Torrent struct {
	PieceHashes [][20]byte
	InfoHash    [20]byte
	PeerId      [20]byte
	Peers       []peers.Peer
	Length      int
	PieceLength int
	Name        string
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
}

type pieceResult struct {
	index int
	buf   []byte
}

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read()
	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

	switch msg.Id {
	case messages.MsgUnchoke:
		state.client.Choked = false
	case messages.MsgChoke:
		state.client.Choked = true
	case messages.MsgPiece:
		n, err := messages.ParsePiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}

		state.downloaded += n
	}

	return nil
}

func (t *Torrent) attemptDownloadPiece(cl *client.Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pw.index,
		client: cl,
		buf:    make([]byte, pw.length),
	}

	cl.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer cl.Conn.SetDeadline(time.Time{})

	for state.downloaded < pw.length {
		if !state.client.Choked {
			for state.requested < pw.length {
				blockSize := MaxBlockSize

				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}

				err := cl.SendRequest(pw.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}

				state.requested += blockSize
			}
		}

		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

func (t *Torrent) calculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("index %d failed integrity check", pw.index)
	}
	return nil
}

func (t *Torrent) startDownloadWorker(peer peers.Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	cl, err := client.New(peer, t.PeerId, t.InfoHash)
	if err != nil {
		log.Println(err)
		return
	}
	defer cl.Conn.Close()

	log.Printf("Complete handshake with %s\n", peer.Ip)

	cl.SendInterested()

	for pw := range workQueue {
		buf, err := t.attemptDownloadPiece(cl, pw)
		if err != nil {
			log.Println(err)
			return
		}

		err = checkIntegrity(pw, buf)
		if err != nil {
			log.Println(err)
			return
		}

		results <- &pieceResult{
			index: pw.index,
			buf:   buf,
		}
	}
}

func (t *Torrent) startDownloadPieceWorker(peer peers.Peer, pw *pieceWork, results chan *pieceResult) {
	cl, err := client.New(peer, t.PeerId, t.InfoHash)
	if err != nil {
		log.Println(err)
		return
	}
	defer cl.Conn.Close()

	cl.SendInterested()

	buf, err := t.attemptDownloadPiece(cl, pw)
	if err != nil {
		log.Println(err)
		return
	}

	err = checkIntegrity(pw, buf)
	if err != nil {
		log.Println(err)
		return
	}

	results <- &pieceResult{
		index: pw.index,
		buf:   buf,
	}
}

func (t *Torrent) DownloadFile() ([]byte, error) {
	log.Printf("Starting download for %s\n", t.Name)

	results := make(chan *pieceResult)
	workQueue := make(chan *pieceWork, len(t.PieceHashes))

	for index, hash := range t.PieceHashes {
		len := t.calculatePieceSize(index)
		workQueue <- &pieceWork{index, hash, len}
	}

	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, results)
	}

	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHashes) {
		res := <-results
		begin, end := t.calculateBoundsForPiece(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++

		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		numWorkers := runtime.NumGoroutine() - 1
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
	}
	close(workQueue)

	return buf, nil
}

func (t *Torrent) DownloadPiece(index int) ([]byte, error) {
	log.Printf("Starting download for %s piece #%d\n", t.Name, index)

	results := make(chan *pieceResult)

	pw := &pieceWork{
		index: index,
		hash:  t.PieceHashes[index],
	}
	pw.length = t.calculatePieceSize(pw.index)

	go t.startDownloadPieceWorker(t.Peers[0], pw, results)

	piece := <-results
	log.Printf("Download piece #%d\n", piece.index)

	return piece.buf, nil
}
