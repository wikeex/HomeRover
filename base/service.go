package base

import (
	"HomeRover/consts"
	"HomeRover/models/client"
	"HomeRover/models/config"
	"HomeRover/models/data"
	"fmt"
	"net"
	"sync"
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
	StateMu		 	sync.RWMutex
}

func (c *ClientService) HoldPunching()  {
	go c.HoldPunching()

	go holdPunchingRecv(c.CmdConn, &c.LocalInfo.Trans.CmdState, &c.StateMu, c.Conf.PackageLen)
	go holdPunchingRecv(c.VideoConn, &c.LocalInfo.Trans.VideoState, &c.StateMu, c.Conf.PackageLen)
	go holdPunchingRecv(c.AudioConn, &c.LocalInfo.Trans.AudioState, &c.StateMu, c.Conf.PackageLen)
}

func (c *ClientService) holdPunchingSend()  {
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

func holdPunchingRecv(conn *net.UDPConn, state *bool, stateMu *sync.RWMutex, packageLen int) {
	recvBytes := make([]byte, packageLen)
	recvData := data.Data{}

	var (
		err error
	)

	for {
		_, _, err = conn.ReadFromUDP(recvBytes)
		if err != nil {
			fmt.Println(err)
		}

		err = recvData.FromBytes(recvBytes)
		if err != nil {
			fmt.Println(err)
		}
		if recvData.Type == consts.HoldPunchingResp {
			stateMu.Lock()
			*state = true
			stateMu.Unlock()
			break
		}
	}
}