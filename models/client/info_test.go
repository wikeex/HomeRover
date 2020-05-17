package client

import (
	"HomeRover/consts"
	"HomeRover/models/mode"
	"net"
	"testing"
)

func TestInfo(t *testing.T) {
	cmdAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	if err != nil {
		t.Error(err)
	}

	videoAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	if err != nil {
		t.Error(err)
	}

	audioAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10002")
	if err != nil {
		t.Error(err)
	}

	trans := mode.Trans{
		Cmd: consts.HoldPunching,
		Video: consts.HoldPunching,
		Audio: consts.HoldPunching,
		CmdState: true,
		VideoState: true,
		AudioState: true,
	}

	info := Info{
		CmdAddr: cmdAddr,
		VideoAddr: videoAddr,
		AudioAddr: audioAddr,
		Id: 1000,
		GroupId: 1000,
		Type: 1,
		Trans: trans,
	}

	infoBytes, err := info.ToBytes()
	if err != nil {
		t.Error(err)
	}

	t.Log(infoBytes)

	info2 := Info{}

	err = info2.FromBytes(infoBytes)
	if err != nil {
		t.Error(err)
	}

	t.Log(info2)
}
