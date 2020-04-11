package controller

import (
	"HomeRover/models/config"
	"HomeRover/models/pack"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"
)

const (
	ServerSend = iota
	SERVER_RECV
	CONTROLLER_SEND
	CONTROLLER_RECV
	VIDEO_SEND
	VIDEO_RECV
	AUDIO_END
	AUDIO_RECV
)

func allocatePort(conn *net.UDPConn) error {
	rand.Seed(time.Now().UnixNano())
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", rand.Intn(55535) + 10000))
	if err != nil {
		return allocatePort(conn)
	}
	conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	return nil
}

type Service struct {
	conf 					*config.ControllerConfig

	serverConn 				*net.UDPConn
	controllerSendConn		*net.UDPConn
	controllerRecvConn		*net.UDPConn
	videoSendConn			*net.UDPConn
	videoRecvConn			*net.UDPConn
	audioSendConn			*net.UDPConn
	audioRecvConn			*net.UDPConn

	roverControllerAddr		*net.UDPAddr
	roverVideoAddr			*net.UDPAddr
	roverAudioAddr			*net.UDPAddr
}

func (s *Service) initConn() error {
	err := allocatePort(s.serverConn)
	if err != nil {
		return err
	}
	err = allocatePort(s.controllerSendConn)
	if err != nil {
		return err
	}
	err = allocatePort(s.controllerRecvConn)
	if err != nil {
		return err
	}
	err = allocatePort(s.videoSendConn)
	if err != nil {
		return err
	}
	err = allocatePort(s.videoRecvConn)
	if err != nil {
		return err
	}
	err = allocatePort(s.audioSendConn)
	if err != nil {
		return err
	}
	err = allocatePort(s.audioRecvConn)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) serverHeartbeat()  {
	var controllerId []byte

	binary.BigEndian.PutUint16(controllerId, uint16(s.conf.ControllerId))
	sendObject := pack.Data{
		Type:    		ServerSend,
		OrderNum:     	0,
		Payload: 		controllerId,
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

func (s *Service) serverRecv(ch chan []byte)  {
	receiveData := make([]byte, s.conf.PackageLen)
	data := pack.Data{}

	for {
		_, _, err := s.serverConn.ReadFromUDP(receiveData)
		if err != nil {
			fmt.Println(err)
		}
		err = data.UnPackage(receiveData)
		if err != nil {
			fmt.Println(err)
		}
		ch <- data.Payload
	}
}