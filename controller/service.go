package controller

import (
	"HomeRover/models/config"
	"HomeRover/models/consts"
	"HomeRover/models/pack"
	"HomeRover/models/server"
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

type Service struct {
	conf 					*config.ControllerConfig

	joystickData			*chan []byte

	serverConn 				*net.UDPConn
	cmdConn					*net.UDPConn
	videoConn				*net.UDPConn
	audioConn				*net.UDPConn

	roverAddr				pack.AddrInfo
	localAddr				pack.AddrInfo
	addrMu					sync.RWMutex
}

func (s *Service) initConn() error {
	_, err := allocatePort(s.serverConn)
	if err != nil {
		return err
	}
	s.localAddr.CmdAddr, err = allocatePort(s.cmdConn)
	if err != nil {
		return err
	}
	s.localAddr.VideoAddr, err = allocatePort(s.videoConn)
	if err != nil {
		return err
	}
	s.localAddr.AudioAddr, err = allocatePort(s.audioConn)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) serverSend()  {
	s.localAddr.Id = uint16(s.conf.ControllerId)
	s.localAddr.Trans = s.conf.Trans

	addrBytes, err := s.localAddr.Package()
	if err != nil {
		fmt.Println(err)
	}
	sendObject := pack.Data{
		Type:    		consts.ControllerServe,
		OrderNum:     	0,
		Payload: 		addrBytes,
	}

	sendData := sendObject.Package()

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
		sendData = sendObject.Package()
	}
}

func (s *Service) serverRecv()  {
	receiveData := make([]byte, s.conf.PackageLen)
	data := pack.Data{}
	rover := server.Client{}

	for {
		_, _, err := s.serverConn.ReadFromUDP(receiveData)
		if err != nil {
			fmt.Println(err)
		}
		err = data.UnPackage(receiveData)
		if err != nil {
			fmt.Println(err)
		}

		if data.Type == consts.ServerResp {
			s.addrMu.Lock()
			err = rover.FromBytes(data.Payload)
			if err != nil {
				fmt.Println(err)
			}
			if rover.State != consts.Online {
				s.addrMu.Unlock()
				continue
			}
			s.roverAddr = rover.Info
			s.addrMu.Unlock()
		}
	}
}

func (s *Service) cmdSend()  {
	sendObject := pack.Data{
		Type:     consts.ControllerCmd,
		OrderNum: 0,
		Payload:  nil,
	}

	var (
		sendData 	[]byte
		err			error
	)

	for {
		sendObject.Payload =  <- *s.joystickData
		sendData = sendObject.Package()
		_, err = s.cmdConn.WriteToUDP(sendData, s.roverAddr.CmdAddr)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (s *Service) cmdRecv() {
	receiveData := make([]byte, s.conf.PackageLen)
	data := pack.Data{}

	for {
		_, _, err := s.cmdConn.ReadFromUDP(receiveData)
		if err != nil {
			fmt.Println(err)
		}
		err = data.UnPackage(receiveData)
		if err != nil {
			fmt.Println(err)
		}

		if data.Type == consts.RoverCmd {
			fmt.Println("rover cmd received")
		}
	}
}

func (s *Service) Run() error {
	err := s.initConn()
	if err != nil {
		return err
	}

	go s.serverSend()
	go s.serverRecv()

	go s.cmdSend()
	go s.cmdRecv()

	return nil
}