package data

import (
	"bytes"
	"encoding/binary"
)

type Data struct {
	Type 		uint8
	OrderNum 	uint16
	Payload		[]byte
}

func (d Data) ToBytes() []byte {
	var buffer bytes.Buffer

	buffer.Write([]byte{d.Type})

	num := make([]byte, 2)
	binary.BigEndian.PutUint16(num, d.OrderNum)
	buffer.Write(num)

	buffer.Write(d.Payload)
	return buffer.Bytes()
}

func (d Data) FromBytes(b []byte) error {
	d.Type = b[0]

	d.OrderNum = binary.BigEndian.Uint16(b[1:3])
	d.Payload = b[3:]

	return nil
}

