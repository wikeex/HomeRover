package controller

import (
	"HomeRover/consts"
	"HomeRover/models/client"
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
		conf:			conf,
		joystickData: 	joystickData,
		roverMu:      	sync.RWMutex{},
		addrMu:       	sync.RWMutex{},
	}

	return
}

type Service struct {
	conf 			*config.ControllerConfig

	joystickData	chan []byte


	serverConn 		*net.UDPConn
	cmdConn   		*net.UDPConn
	videoConn  		*net.UDPConn
	audioConn  		*net.UDPConn

	rover     	client.Client
	localInfo 		client.Info
	roverMu   		sync.RWMutex
	addrMu    		sync.RWMutex
}

func (s *Service) initConn() error {
	_, err := allocatePort(s.serverConn)
	if err != nil {
		return err
	}
	s.localInfo.CmdAddr, err = allocatePort(s.cmdConn)
	if err != nil {
		return err
	}
	s.localInfo.VideoAddr, err = allocatePort(s.videoConn)
	if err != nil {
		return err
	}
	s.localInfo.AudioAddr, err = allocatePort(s.audioConn)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) serverSend()  {
	s.localInfo.Id = uint16(s.conf.ControllerId)
	s.localInfo.Trans = s.conf.Trans

	addrBytes, err := s.localInfo.ToBytes()
	if err != nil {
		fmt.Println(err)
	}
	sendObject := data.Data{
		Type:     consts.ControllerServe,
		OrderNum: 0,
		Payload:  addrBytes,
	}

	sendData := sendObject.ToBytes()

	addrStr := s.conf.ServerIP + ":" + strconv.Itoa(s.conf.ServerPort)
	addr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		fmt.Println(err)
	}

	for range time.Tick(time.Second){
		_, err = s.serverConn.WriteToUDP(sendData, addr)
		if err != nil {
			fmt.Println(err)
		}
		sendObject.OrderNum++
		sendData = sendObject.ToBytes()
	}
}

func (s *Service) serverRecv()  {
	receiveData := make([]byte, s.conf.PackageLen)
	RecvData := data.Data{}

	for {
		_, _, err := s.serverConn.ReadFromUDP(receiveData)
		if err != nil {
			fmt.Println(err)
		}
		err = RecvData.FromBytes(receiveData)
		if err != nil {
			fmt.Println(err)
		}

		if RecvData.Type == consts.ServerResp {
			s.roverMu.Lock()
			err = s.rover.FromBytes(RecvData.Payload)
			if err != nil {
				fmt.Println(err)
			}
			s.roverMu.Unlock()
			if s.rover.State == consts.Offline {
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
		GroupId: s.rover.Info.GroupId,
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
		if s.rover.State == consts.Online {
			sendData = sendObject.ToBytes()
			_, err = s.cmdConn.WriteToUDP(sendData, s.rover.Info.CmdAddr)
			if err != nil {
				fmt.Println(err)
			}
		}
		s.roverMu.RUnlock()
	}
}

func (s *Service) cmdRecv() {
	receiveData := make([]byte, s.conf.PackageLen)
	recvData := data.Data{}

	for {
		_, _, err := s.cmdConn.ReadFromUDP(receiveData)
		if err != nil {
			fmt.Println(err)
		}
		err = recvData.FromBytes(receiveData)
		if err != nil {
			fmt.Println(err)
		}

		if recvData.Type == consts.RoverCmd {
			fmt.Println("rover cmd received")
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