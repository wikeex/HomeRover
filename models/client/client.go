package client

import (
	"bytes"
)

type State uint8

// Controller and Rover are all Client
type Client struct {
	State State
	Info  Info
}

func (c Client) ToBytes() ([]byte, error) {
	buffer := bytes.Buffer{}

	buffer.Write([]byte{byte(c.State)})

	roverInfoBytes, err := c.Info.ToBytes()
	if err != nil {
		return nil, err
	}

	buffer.Write(roverInfoBytes)

	return buffer.Bytes(), nil
}

func (c Client) FromBytes(b []byte) error {
	c.State = State(b[0])

	err := c.Info.FromBytes(b[1:])
	if err != nil {
		return err
	}
	return nil
}

