package base

import (
	"HomeRover/consts"
	"HomeRover/models/client"
	"HomeRover/models/config"
	"HomeRover/models/data"
	"fmt"
	"net"
	"time"
)

type ClientService struct {
	Conf 			*config.ControllerConfig

	ServerConn 		*net.UDPConn
	CmdConn   		*net.UDPConn
	VideoConn  		*net.UDPConn
	AudioConn  		*net.UDPConn

	DestClient     	client.Client
	LocalInfo 		client.Info
}

func (c *ClientService) HoldPunching()  {
	var (
		sendData	data.Data
		err			error
	)

	sendData.Type = consts.HoldPunchingSend
	for range time.Tick(time.Second) {
		if c.DestClient.Info.Trans.Cmd == consts.HoldPunching {
			_, err = c.CmdConn.WriteToUDP(sendData.ToBytes(), c.DestClient.Info.CmdAddr)
			if err != nil {
				fmt.Println(err)
			}
		}

		if c.DestClient.Info.Trans.Video == consts.HoldPunching {
			_, err = c.CmdConn.WriteToUDP(sendData.ToBytes(), c.DestClient.Info.VideoAddr)
			if err != nil {
				fmt.Println(err)
			}
		}

		if c.DestClient.Info.Trans.Audio == consts.HoldPunching {
			_, err = c.CmdConn.WriteToUDP(sendData.ToBytes(), c.DestClient.Info.AudioAddr)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}