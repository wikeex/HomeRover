package client

import (
	"HomeRover/models/consts"
	"HomeRover/models/mode"
	"bytes"
	"encoding/binary"
	"net"
	"strconv"
	"strings"
)

type Info struct {
	Id					uint16

	// every addr takes 6 Bytes, 4 Bytes for IP, 2 Bytes for Port
	CmdAddr				*net.UDPAddr
	VideoAddr			*net.UDPAddr
	AudioAddr			*net.UDPAddr

	GroupId				uint16
	Trans				mode.Trans
}

func (c Info) ToBytes() ([]byte, error) {
	var buffer bytes.Buffer

	// client client id
	idBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(idBytes, c.Id)
	buffer.Write(idBytes)

	// client addr
	cmdBytes, err := infoToBytes(c.CmdAddr)
	if err != nil {
		return nil, err
	}

	videoBytes, err := infoToBytes(c.VideoAddr)
	if err != nil {
		return nil, err
	}

	audioBytes, err := infoToBytes(c.AudioAddr)
	if err != nil {
		return nil, err
	}
	buffer.Write(cmdBytes)
	buffer.Write(videoBytes)
	buffer.Write(audioBytes)

	// client group id
	groupIdBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(groupIdBytes, c.GroupId)
	buffer.Write(groupIdBytes)

	// client trans mode to buffer
	// set modeByte lowest three bits to configure to hold punching mode
	var modeByte uint8 = 0
	if c.Trans.Cmd == consts.HoldPunching {
		modeByte |= 1 << 2
	} else {
		modeByte |= 0 << 2
	}

	if c.Trans.Video == consts.HoldPunching {
		modeByte |= 1 << 1
	} else {
		modeByte |= 0 << 1
	}

	if c.Trans.Audio == consts.HoldPunching {
		modeByte |= 1
	} else {
		modeByte |= 0
	}

	buffer.Write([]byte{modeByte})

	return buffer.Bytes(), nil
}

func (c Info) FromBytes(b []byte) error {
	var err error
	c.Id = binary.BigEndian.Uint16(b[:2])

	c.CmdAddr, err = bytesToInfo(b[2:8])
	if err != nil {
		return err
	}

	c.VideoAddr, err = bytesToInfo(b[8:14])
	if err != nil {
		return err
	}

	c.AudioAddr, err = bytesToInfo(b[14:20])
	if err != nil {
		return err
	}

	c.GroupId = binary.BigEndian.Uint16(b[20:22])

	modeByte := b[22]
	if modeByte & 1 << 2 == 4 {
		c.Trans.Cmd = consts.HoldPunching
	} else {
		c.Trans.Cmd = consts.ServerForwarding
	}

	if modeByte & 1 << 1 == 2 {
		c.Trans.Video = consts.HoldPunching
	} else {
		c.Trans.Video = consts.ServerForwarding
	}

	if modeByte & 1 == 1 {
		c.Trans.Audio = consts.HoldPunching
	} else {
		c.Trans.Audio = consts.ServerForwarding
	}

	return nil
}

func bytesToInfo(b []byte) (*net.UDPAddr, error) {
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

func infoToBytes(addr *net.UDPAddr) ([]byte, error) {
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
