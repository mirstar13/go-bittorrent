package handshake

import (
	"bytes"
	"fmt"
	"io"
)

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerId   [20]byte
}

func New(infoHash, peerId [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: infoHash,
		PeerId:   peerId,
	}
}

func (hs *Handshake) Serialize() *bytes.Buffer {
	var buf bytes.Buffer
	buf.WriteByte(byte(len([]byte(hs.Pstr))))
	buf.Write([]byte(hs.Pstr))
	buf.Write(make([]byte, 8))
	buf.Write(hs.InfoHash[:])
	buf.Write(hs.PeerId[:])
	return &buf
}

func Read(r io.Reader) (*Handshake, error) {
	pstrBuf := make([]byte, 1)
	_, err := io.ReadFull(r, pstrBuf)
	if err != nil {
		return nil, err
	}

	pstrLen := int(pstrBuf[0])
	if pstrLen == 0 {
		return nil, fmt.Errorf("pstr length can not be 0")
	}

	handshakeBuf := make([]byte, 48+pstrLen)
	_, err = io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, err
	}

	reservedBytes := 8
	var infoHash, peerId [20]byte

	copy(infoHash[:], handshakeBuf[reservedBytes+pstrLen:reservedBytes+pstrLen+20])
	copy(peerId[:], handshakeBuf[reservedBytes+pstrLen+20:])

	h := Handshake{
		Pstr:     string(handshakeBuf[:pstrLen]),
		InfoHash: infoHash,
		PeerId:   peerId,
	}

	return &h, nil
}
