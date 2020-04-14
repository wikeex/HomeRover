package server

import (
	"HomeRover/models/pack"
	"bytes"
)

type ClientState uint8

// Controller and Rover are all Client
type Client struct {
	State			ClientState
	Info 			pack.AddrInfo
}

func (c Client) ToBytes() ([]byte, error) {
	buffer := bytes.Buffer{}

	buffer.Write([]byte{byte(c.State)})

	roverInfoBytes, err := c.Info.Package()
	if err != nil {
		return nil, err
	}

	buffer.Write(roverInfoBytes)

	return buffer.Bytes(), nil
}

func (c Client) FromBytes(b []byte) error {
	c.State = ClientState(b[0])

	err := c.Info.UnPackage(b[1:])
	if err != nil {
		return err
	}
	return nil
}

