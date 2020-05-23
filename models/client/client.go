package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type State uint8

// Controller and Rover are all Client
type Client struct {
	Id 			uint16
	GroupId		uint16
	State 		State
	Type 		uint8
	Payload 	[]byte

	// addr and sendCh will not to bytes
	Addr		*net.Addr
	SendCh		chan []byte
}

func (c *Client) ToBytes() ([]byte, error) {
	buffer := bytes.Buffer{}

	idBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(idBytes, c.Id)

	groupIdBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(groupIdBytes, c.GroupId)

	buffer.Write(idBytes)
	buffer.Write(groupIdBytes)

	buffer.Write([]byte{byte(c.State)})

	buffer.Write([]byte{c.Type})

	buffer.Write(c.Payload)

	return buffer.Bytes(), nil
}

func (c *Client) FromBytes(b []byte) error {
	if len(b) <= 0 {
		return fmt.Errorf("bytes is empty")
	}
	c.Id = binary.BigEndian.Uint16(b[:2])
	c.GroupId = binary.BigEndian.Uint16(b[2:4])

	c.State = State(b[4])

	c.Type = b[5]

	c.Payload = b[6:]

	return nil
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
	var portBytes []byte
	binary.BigEndian.PutUint16(portBytes, uint16(numInt))

	buffer.Write(portBytes)
	return buffer.Bytes(), nil
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