package client

import (
	"HomeRover/consts"
	"HomeRover/models/mode"
	"bytes"
	"encoding/binary"
	"strconv"
	"strings"
)

type Info struct {
	Id					uint16

	CmdPort				uint16
	VideoPort			uint16
	AudioPort			uint16

	IP					string

	GroupId				uint16
	Trans				mode.Trans

	Type 				uint8
}

func (c *Info) ToBytes() ([]byte, error) {
	var buffer bytes.Buffer

	// client client id
	idBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(idBytes, c.Id)
	buffer.Write(idBytes)

	// client addr
	cmdBytes, err := portToBytes(c.CmdPort)
	if err != nil {
		return nil, err
	}

	videoBytes, err := portToBytes(c.VideoPort)
	if err != nil {
		return nil, err
	}

	audioBytes, err := portToBytes(c.AudioPort)
	if err != nil {
		return nil, err
	}
	buffer.Write(cmdBytes)
	buffer.Write(videoBytes)
	buffer.Write(audioBytes)

	// ip
	if c.IP == "" {
		c.IP = "0.0.0.0"
	}
	ipStrings := strings.Split(c.IP, ".")
	for _, item := range ipStrings {
		numInt, err := strconv.Atoi(item)
		if err != nil {
			return nil, err
		}
		buffer.Write([]byte{uint8(numInt)})
	}

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

	buffer.Write([]byte{c.Type})

	return buffer.Bytes(), nil
}

func (c *Info) FromBytes(b []byte) error {
	c.Id = binary.BigEndian.Uint16(b[:2])

	c.CmdPort = binary.BigEndian.Uint16(b[2:4])
	c.VideoPort = binary.BigEndian.Uint16(b[4:6])
	c.AudioPort = binary.BigEndian.Uint16(b[6:8])

	var ipStrings []string
	for _, num := range b[8:12] {
		ipStrings = append(ipStrings, strconv.Itoa(int(num)))
	}
	c.IP = strings.Join(ipStrings, ".")

	c.GroupId = binary.BigEndian.Uint16(b[12:14])

	modeByte := b[14]
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

	c.Type = b[15]

	return nil
}

func portToBytes(port uint16) ([]byte, error) {
	var buffer bytes.Buffer

	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, port)
	buffer.Write(portBytes)
	return buffer.Bytes(), nil
}
