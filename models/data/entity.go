package data

import (
	"bytes"
	"encoding/binary"
)

type EntityData struct {
	GroupId		uint16
	Payload		[]byte
}

func (e EntityData) ToBytes() []byte {
	var buffer bytes.Buffer

	groupIdBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(groupIdBytes, e.GroupId)
	buffer.Write(groupIdBytes)

	buffer.Write(e.Payload)
	return buffer.Bytes()
}

func (e EntityData) FromBytes(b []byte) error {
	e.GroupId = binary.BigEndian.Uint16(b[:2])
	e.Payload = b[2:]

	return nil
}