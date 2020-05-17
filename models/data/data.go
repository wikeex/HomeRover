package data

import (
	"bytes"
	"encoding/binary"
)

type Data struct {
	Type 		uint8
	Channel		uint8
	OrderNum 	uint16
	Payload		[]byte
}

func (d *Data) ToBytes() []byte {
	var buffer bytes.Buffer

	buffer.WriteByte(d.Type)
	buffer.WriteByte(d.Channel)
	num := make([]byte, 2)
	binary.BigEndian.PutUint16(num, d.OrderNum)
	buffer.Write(num)

	buffer.Write(d.Payload)
	return buffer.Bytes()
}

func (d *Data) FromBytes(b []byte) error {
	d.Type = b[0]
	d.Channel = b[1]
	d.OrderNum = binary.BigEndian.Uint16(b[2:4])
	d.Payload = b[4:]

	return nil
}