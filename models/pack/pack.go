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
	Id				uint16

	// every addr takes 6 Bytes, 4 Bytes for IP, 2 Bytes for Port
	CmdAddr				*net.UDPAddr
	VideoAddr			*net.UDPAddr
	AudioAddr			*net.UDPAddr
}

func (a AddrInfo) Package() ([]byte, error) {
	var buffer bytes.Buffer

	destIdBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(destIdBytes, a.Id)
	buffer.Write(destIdBytes)

	cmdBytes, err := addrToBytes(a.CmdAddr)
	if err != nil {
		return nil, err
	}

	videoBytes, err := addrToBytes(a.VideoAddr)
	if err != nil {
		return nil, err
	}

	audioBytes, err := addrToBytes(a.AudioAddr)
	if err != nil {
		return nil, err
	}

	buffer.Write(cmdBytes)
	buffer.Write(videoBytes)
	buffer.Write(audioBytes)

	return buffer.Bytes(), nil
}

func (a AddrInfo) UnPackage(b []byte) error {
	var err error
	a.Id = binary.BigEndian.Uint16(b[:2])

	a.CmdAddr, err = bytesToAddr(b[2:8])
	if err != nil {
		return err
	}

	a.VideoAddr, err = bytesToAddr(b[8:14])
	if err != nil {
		return err
	}

	a.AudioAddr, err = bytesToAddr(b[14:20])
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
