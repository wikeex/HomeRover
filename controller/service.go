package controller

import (
	"HomeRover/base"
	"HomeRover/consts"
	"HomeRover/models/config"
	"HomeRover/models/data"
	"fmt"
	"sync"
)


func NewService(conf *config.CommonConfig, controllerConf *config.ControllerConfig, joystickData chan []byte) (service *Service, err error) {
	service = &Service{
		joystickData: 	joystickData,
		infoMu:       	sync.RWMutex{},
	}

	service.Conf = conf
	service.controllerConf = controllerConf
	service.LocalInfo.Type = consts.Controller
	service.LocalInfo.Id = uint16(conf.Id)
	service.LocalInfo.Trans = controllerConf.Trans
	return
}

type Service struct {
	base.Service

	controllerConf 	*config.ControllerConfig

	joystickData	chan []byte

	infoMu    		sync.RWMutex
}

func (s *Service) cmdSend()  {
	sendObject := data.Data{
		Type:     consts.Controller,
		Channel:  consts.Cmd,
		OrderNum: 0,
		Payload:  nil,
	}

	sendEntity := data.EntityData{
		GroupId: s.DestClient.Info.GroupId,
		Payload: nil,
	}

	var (
		sendData 	[]byte
		err			error
	)

	for {
		sendEntity.Payload =  <- s.joystickData
		sendObject.Payload = sendEntity.ToBytes()
		s.DestClientMu.RLock()
		if s.DestClient.State == consts.Online {
			sendData = sendObject.ToBytes()
			_, err = s.CmdConn.WriteToUDP(sendData, s.DestClient.Info.CmdAddr)
			if err != nil {
				fmt.Println(err)
			}
		}
		s.DestClientMu.RUnlock()
	}
}

func (s *Service) cmdRecv() {
	recvBytes := make([]byte, s.Conf.PackageLen)
	recvData := data.Data{}

	for {
		_, _, err := s.CmdConn.ReadFromUDP(recvBytes)
		if err != nil {
			fmt.Println(err)
		}
		err = recvData.FromBytes(recvBytes)
		if err != nil {
			fmt.Println(err)
		}

		if recvData.Type == consts.Rover && recvData.Channel == consts.Cmd {
			fmt.Println("cmd received")
		}
	}
}

func (s *Service) Run() {
	err := s.InitConn()
	if err != nil {
		fmt.Println(err)
	}

	go s.ServerSend()
	go s.ServerRecv()

	go s.cmdSend()
	go s.cmdRecv()
}