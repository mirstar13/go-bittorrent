package peers

import (
	"encoding/binary"
	"net"
	"strconv"
)

type Peer struct {
	Ip   net.IP
	Port uint16
}

func Unmarshal(rawPeers []byte) []Peer {
	resultPeers := []Peer{}
	for i := 0; i < len(rawPeers); i += 6 {
		port := binary.BigEndian.Uint16([]byte(rawPeers[i+4 : i+6]))

		peer := Peer{
			Ip:   net.IP(rawPeers[i : i+4]),
			Port: uint16(port),
		}

		resultPeers = append(resultPeers, peer)
	}

	return resultPeers
}

func (p *Peer) StringAddr() string {
	return net.JoinHostPort(p.Ip.String(), strconv.Itoa(int(p.Port)))
}
