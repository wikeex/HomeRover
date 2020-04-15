package server

import (
	"HomeRover/models/client"
	"HomeRover/models/config"
	"HomeRover/models/consts"
	"HomeRover/models/data"
	"HomeRover/models/server"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"net"
	"sync"
)

type Service struct {
	conf 				*config.ControllerConfig
	confMu				sync.RWMutex

	Groups				map[uint16]*server.Group
	TransMu				sync.RWMutex
	clientMu			sync.RWMutex

	clientConn 			*net.UDPConn
}

func (s *Service) init() error {
	groupSet := mapset.NewSet()
	groupSet.Add(uint16(1))
	groupSet.Add(uint16(2))
	s.Groups[0] = &server.Group{
		Id: 0,
		Rover: client.Client{
			Info: client.Info{Id: 1},
		},
		Controller: client.Client{
			Info: client.Info{Id: 2},
		},
	}

	s.confMu.RLock()
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", s.conf.LocalPort))
	if err != nil {
		return err
	}
	s.confMu.RUnlock()

	s.clientConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	
	return nil
}

func (s *Service)listenClients()  {
	receiveData := make([]byte, s.conf.PackageLen)
	recvData := data.Data{}
	var (
		err        error
		addr       *net.UDPAddr
		ClientInfo client.Info
	)

	for {
		_, addr, err = s.clientConn.ReadFromUDP(receiveData)
		if err != nil {
			fmt.Println(err)
		}
		err = recvData.FromBytes(receiveData)
		if err != nil {
			fmt.Println(err)
		}

		if recvData.Type == consts.ControllerServe || recvData.Type == consts.RoverServe {
			fmt.Println("heartbeat received")
			err = ClientInfo.FromBytes(recvData.Payload)
			if err != nil {
				fmt.Println(err)
			}

			if recvData.Type == consts.ControllerServe {
				// get the dest client from s.Groups
				s.clientMu.Lock()
				s.Groups[ClientInfo.GroupId].Controller.Info = ClientInfo
				s.Groups[ClientInfo.GroupId].Controller.State = consts.Online
				s.clientMu.Unlock()

				s.TransMu.Lock()
				s.Groups[ClientInfo.GroupId].Trans = &ClientInfo.Trans
				s.TransMu.Unlock()

				// send rover addr back
				recvData.Type = consts.ServerResp

				recvData.Payload, err = s.Groups[ClientInfo.GroupId].Rover.ToBytes()
				if err != nil {
					fmt.Println(err)
				}
			} else if recvData.Type == consts.RoverServe {
				// get the dest client from s.Groups
				s.clientMu.Lock()
				s.Groups[ClientInfo.GroupId].Rover.Info = ClientInfo
				s.Groups[ClientInfo.GroupId].Rover.State = consts.Online
				s.clientMu.Unlock()

				// send controller addr back
				recvData.Type = consts.ServerResp

				recvData.Payload, err = s.Groups[ClientInfo.GroupId].Rover.ToBytes()
				if err != nil {
					fmt.Println(err)
				}
			}

			_, err = s.clientConn.WriteTo(recvData.ToBytes(), addr)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (s *Service) Run() error {
	err := s.init()
	if err != nil {
		return err
	}
	
	go s.listenClients()

	return nil
}