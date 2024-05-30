package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/mirstar13/go-bittorrent/client"
	"github.com/mirstar13/go-bittorrent/handshake"
	"github.com/mirstar13/go-bittorrent/parser"
	"github.com/mirstar13/go-bittorrent/torrent"
)

func main() {
	log.Println("Logs from your program will appear here!")

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)

	case "decode":
		bencodedValue := args[0]
		decoded, _, err := parser.Decode(bencodedValue)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))

	case "info":
		torrFileName := args[0]

		torrFile, err := torrent.Open(torrFileName)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		pieceHashes := ""
		for _, pieceHash := range torrFile.PieceHashes {
			pieceHashes += hex.EncodeToString(pieceHash[:]) + "\n"
		}
		infoHash := hex.EncodeToString(torrFile.InfoHash[:])

		fmt.Printf("Tracker URL: %s\nLength: %d\nInfo Hash: %s\nPiece Length: %d\nPiece Hashes:\n%s",
			torrFile.Announce,
			torrFile.Length,
			infoHash,
			torrFile.PieceLength,
			pieceHashes,
		)
	case "peers":
		torrFileName := args[0]

		torrFile, err := torrent.Open(torrFileName)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		torPeers, err := torrFile.RequestPeers(torrent.PeerId, torrent.Port)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		for _, peer := range torPeers {
			fmt.Printf("%s:%d\n", peer.Ip, peer.Port)
		}
	case "handshake":
		torrFileName := args[0]
		peerAddrs := args[1]

		torrFile, err := torrent.Open(torrFileName)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		h := handshake.New(torrFile.InfoHash, torrent.PeerId)

		c, err := net.Dial("tcp", peerAddrs)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		resp, err := client.CompleteHandshake(c, h.InfoHash, h.PeerId)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Peer ID: %s\n", hex.EncodeToString(resp.PeerId[:]))
	case "download_piece":
		dir := args[1]
		torrFileName := args[2]
		strIndex := args[3]

		index, err := strconv.Atoi(strIndex)
		if err != nil {
			log.Printf("Invalid piece index: %s", err)
			os.Exit(1)
		}

		torr, err := torrent.Open(torrFileName)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		if err := torr.DownloadPiece(index, dir); err != nil {
			log.Println(err)
			os.Exit(1)
		}
	case "download":
		dir := args[1]
		torrFileName := args[2]

		torr, err := torrent.Open(torrFileName)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		if err := torr.DownloadFile(dir); err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}
}
