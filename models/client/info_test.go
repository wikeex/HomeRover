package client

import (
	"HomeRover/consts"
	"HomeRover/models/mode"
	"testing"
)

func TestInfo(t *testing.T) {

	trans := mode.Trans{
		Cmd: consts.ServerForwarding,
		Video: consts.HoldPunching,
		Audio: consts.HoldPunching,
		CmdState: true,
		VideoState: true,
		AudioState: true,
	}

	info := Info{
		CmdPort: 10000,
		VideoPort: 10001,
		AudioPort: 10002,
		IP: "127.0.0.1",
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

	if info.Trans.Cmd {
		t.Log("pass")
	}
}
