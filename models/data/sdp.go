package data

import (
	"bytes"
	"encoding/json"
	"github.com/pion/webrtc/v2"
)

type SDPData struct {
	Type		uint8
	SDPInfo		webrtc.SessionDescription
}

func (s *SDPData) ToBytes() ([]byte, error) {

	buffer := bytes.Buffer{}

	buffer.Write([]byte{s.Type})
	b, err := json.Marshal(s.SDPInfo)
	if err != nil {
		return nil, err
	}
	buffer.Write(b)
	return buffer.Bytes(), nil
}

func (s *SDPData) FromBytes(b []byte) error {
	s.Type = b[0]
	err := json.Unmarshal(b[1:], &s.SDPInfo)
	if err != nil {
		return err
	}
	return nil
}
