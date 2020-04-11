package pack

import (
	"bytes"
	"encoding/binary"
	"net"
	"strconv"
	"strings"
)

type Data struct {
	Type 		uint8
	OrderNum 	uint16
	Payload		[]byte
}

func (d Data) Package() []byte {
	var buffer bytes.Buffer

	buffer.Write([]byte{d.Type})

	num := make([]byte, 2)
	binary.BigEndian.PutUint16(num, d.OrderNum)
	buffer.Write(num)

	buffer.Write(d.Payload)
	return buffer.Bytes()
}

func (d Data) UnPackage(b []byte) error {
	d.Type = b[0]

	d.OrderNum = binary.BigEndian.Uint16(b[1:3])
	d.Payload = b[3:]

	return nil
}


type AddrInfo struct {
	DestId				uint16

	// every addr takes 6 Bytes, 4 Bytes for IP, 2 Bytes for Port
	Addr				*net.UDPAddr
	VideoAddr			*net.UDPAddr
	AudioAddr			*net.UDPAddr
}

func (a AddrInfo) Package() ([]byte, error) {
	var buffer bytes.Buffer

	destIdBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(destIdBytes, a.DestId)
	buffer.Write(destIdBytes)

	addrBytes, err := addrToBytes(a.Addr)
	if err != nil {
		return nil, err
	}
	videoAddrBytes, err := addrToBytes(a.VideoAddr)
	if err != nil {
		return nil, err
	}
	audioAddrBytes, err := addrToBytes(a.AudioAddr)
	if err != nil {
		return nil, err
	}

	buffer.Write(addrBytes)
	buffer.Write(videoAddrBytes)
	buffer.Write(audioAddrBytes)

	return buffer.Bytes(), nil
}

func (a AddrInfo) UnPackage(b []byte) error {
	var err error
	a.Addr, err = bytesToAddr(b[:6])
	if err != nil {
		return err
	}

	a.VideoAddr, err = bytesToAddr(b[6:12])
	if err != nil {
		return err
	}

	a.AudioAddr, err = bytesToAddr(b[12:18])
	if err != nil {
		return err
	}
	return nil
}

func bytesToAddr(b []byte) (*net.UDPAddr, error) {
	var ipStrings []string
	for _, num := range b[:4] {
		ipStrings = append(ipStrings, strconv.Itoa(int(num)))
	}
	ip := strings.Join(ipStrings, ".")
	port := strconv.Itoa(int(binary.BigEndian.Uint16(b[4:])))
	addr, err := net.ResolveUDPAddr("udp", strings.Join([]string{ip, port}, ";"))
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func addrToBytes(addr *net.UDPAddr) ([]byte, error) {
	var buffer bytes.Buffer
	tempString := strings.Split((*addr).String(), ":")
	ipStrings := strings.Split(tempString[0], ".")
	for _, item := range ipStrings {
		numInt, err := strconv.Atoi(item)
		if err != nil {
			return nil, err
		}
		buffer.Write([]byte{uint8(numInt)})
	}
	numInt, err := strconv.Atoi(tempString[1])
	if err != nil {
		return nil, err
	}
	buffer.Write([]byte{uint8(numInt)})
	return buffer.Bytes(), nil
}
