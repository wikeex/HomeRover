package client

import (
	"bytes"
	"fmt"
)

type State uint8

// Controller and Rover are all Client
type Client struct {
	State State
	Info  Info
}

func (c *Client) ToBytes() ([]byte, error) {
	buffer := bytes.Buffer{}

	buffer.Write([]byte{byte(c.State)})

	roverInfoBytes, err := c.Info.ToBytes()
	if err != nil {
		return nil, err
	}

	buffer.Write(roverInfoBytes)

	return buffer.Bytes(), nil
}

func (c *Client) FromBytes(b []byte) error {
	if len(b) <= 0 {
		return fmt.Errorf("bytes is empty")
	}
	c.State = State(b[0])

	err := c.Info.FromBytes(b[1:])
	if err != nil {
		return err
	}
	return nil
}

