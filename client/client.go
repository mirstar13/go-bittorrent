package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/mirstar13/go-bittorrent/handshake"
	"github.com/mirstar13/go-bittorrent/messages"
	"github.com/mirstar13/go-bittorrent/peers"
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	peer     peers.Peer
	infoHash [20]byte
	peerId   [20]byte
}

func CompleteHandshake(conn net.Conn, infohash, peerID [20]byte) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	req := handshake.New(infohash, peerID)
	_, err := conn.Write(req.Serialize().Bytes())
	if err != nil {
		return nil, err
	}

	res, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(res.InfoHash[:], infohash[:]) {
		return nil, fmt.Errorf("expected infohash %x but got %x", res.InfoHash, infohash)
	}
	return res, nil
}

func New(peer peers.Peer, peerId, infoHash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.StringAddr(), 3*time.Second)
	if err != nil {
		return nil, err
	}

	_, err = CompleteHandshake(conn, infoHash, peerId)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   true,
		peer:     peer,
		infoHash: infoHash,
		peerId:   peerId,
	}, nil
}

func (c *Client) Read() (*messages.Message, error) {
	msg, err := messages.Read(c.Conn)
	return msg, err
}

func (c *Client) SendRequest(index, begin, length int) error {
	req := messages.FormatRequest(index, begin, length)
	_, err := c.Conn.Write(req.Serialize())
	return err
}

func (c *Client) SendInterested() error {
	msg := messages.Message{Id: messages.MsgInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendUnchoke() error {
	msg := messages.Message{Id: messages.MsgUnchoke}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}
