package controller

import (
	"HomeRover/base"
	"HomeRover/consts"
	"HomeRover/models/config"
	"HomeRover/models/data"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

func allocatePort(conn *net.UDPConn) (*net.UDPAddr, error) {
	rand.Seed(time.Now().UnixNano())
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", rand.Intn(55535) + 10000))
	if err != nil {
		return allocatePort(conn)
	}
	conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func NewService(conf *config.ControllerConfig, joystickData chan []byte) (service *Service, err error) {
	service = &Service{
		joystickData: joystickData,
		roverMu:      sync.RWMutex{},
		addrMu:       sync.RWMutex{},
	}

	service.Conf = conf
	return
}

type Service struct {
	base.ClientService

	joystickData			chan []byte

	roverMu   				sync.RWMutex
	addrMu    				sync.RWMutex
}

func (s *Service) initConn() error {
	_, err := allocatePort(s.ServerConn)
	if err != nil {
		return err
	}
	s.LocalInfo.CmdAddr, err = allocatePort(s.CmdConn)
	if err != nil {
		return err
	}
	s.LocalInfo.VideoAddr, err = allocatePort(s.VideoConn)
	if err != nil {
		return err
	}
	s.LocalInfo.AudioAddr, err = allocatePort(s.AudioConn)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) serverSend()  {
	s.LocalInfo.Id = uint16(s.Conf.ControllerId)
	s.LocalInfo.Trans = s.Conf.Trans

	addrBytes, err := s.LocalInfo.ToBytes()
	if err != nil {
		fmt.Println(err)
	}
	sendObject := data.Data{
		Type:     consts.ControllerServe,
		OrderNum: 0,
		Payload:  addrBytes,
	}

	sendData := sendObject.ToBytes()

	addrStr := s.Conf.ServerIP + ":" + strconv.Itoa(s.Conf.ServerPort)
	addr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		fmt.Println(err)
	}

	for range time.Tick(time.Second){
		_, err = s.ServerConn.WriteToUDP(sendData, addr)
		if err != nil {
			fmt.Println(err)
		}
		sendObject.OrderNum++
		sendData = sendObject.ToBytes()
	}
}

func (s *Service) serverRecv()  {
	receiveData := make([]byte, s.Conf.PackageLen)
	RecvData := data.Data{}

	for {
		_, _, err := s.ServerConn.ReadFromUDP(receiveData)
		if err != nil {
			fmt.Println(err)
		}
		err = RecvData.FromBytes(receiveData)
		if err != nil {
			fmt.Println(err)
		}

		if RecvData.Type == consts.ServerResp {
			s.roverMu.Lock()
			err = s.DestClient.FromBytes(RecvData.Payload)
			if err != nil {
				fmt.Println(err)
			}
			s.roverMu.Unlock()
			if s.DestClient.State == consts.Offline {
				fmt.Println("rover is offline")
			}
		}
	}
}

func (s *Service) cmdSend()  {
	sendObject := data.Data{
		Type:     consts.ControllerCmd,
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
		s.roverMu.RLock()
		if s.DestClient.State == consts.Online &&
			((s.DestClient.Info.Trans.CmdState &&
				s.DestClient.Info.Trans.Cmd == consts.HoldPunching) ||
					s.DestClient.Info.Trans.Cmd == consts.ServerForwarding) {
			sendData = sendObject.ToBytes()
			_, err = s.CmdConn.WriteToUDP(sendData, s.DestClient.Info.CmdAddr)
			if err != nil {
				fmt.Println(err)
			}
		}
		s.roverMu.RUnlock()
	}
}

func (s *Service) cmdRecv() {
	recvBytes := make([]byte, s.Conf.PackageLen)
	recvData := data.Data{}
	sendData := data.Data{}

	for {
		_, _, err := s.CmdConn.ReadFromUDP(recvBytes)
		if err != nil {
			fmt.Println(err)
		}
		err = recvData.FromBytes(recvBytes)
		if err != nil {
			fmt.Println(err)
		}

		if recvData.Type == consts.RoverCmd {
			fmt.Println("rover cmd received")
		} else if recvData.Type == consts.HoldPunchingSend {
			_, err = s.CmdConn.WriteToUDP(sendData.ToBytes(), s.DestClient.Info.CmdAddr)
			if err != nil {
				fmt.Println(err)
			}
		} else if recvData.Type == consts.HoldPunchingResp {
			s.roverMu.Lock()
			s.DestClient.Info.Trans.CmdState = true
			s.roverMu.Unlock()
		}
	}
}

func (s *Service) Run() {
	err := s.initConn()
	if err != nil {
		fmt.Println(err)
	}

	go s.serverSend()
	go s.serverRecv()

	go s.cmdSend()
	go s.cmdRecv()
}