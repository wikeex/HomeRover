package server

import (
	"HomeRover/models/config"
	"HomeRover/models/consts"
	"HomeRover/models/group"
	"HomeRover/models/pack"
	"HomeRover/models/server"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"net"
	"sync"
)

type Service struct {
	conf 				*config.ControllerConfig
	confMu				sync.RWMutex

	Groups				map[uint16]*group.Group
	TransMu				sync.RWMutex
	clientMu			sync.RWMutex

	clientConn 			*net.UDPConn
}

func (s *Service) init() error {
	groupSet := mapset.NewSet()
	groupSet.Add(uint16(1))
	groupSet.Add(uint16(2))
	s.Groups[0] = &group.Group{
		Id: 0,
		Rover: server.Client{
			DestInfo: pack.AddrInfo{Id: 1},
		},
		Controller: server.Client{
			DestInfo: pack.AddrInfo{Id: 2},
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
	data := pack.Data{}
	var (
		err				error
		addr 			*net.UDPAddr
		addrInfo		pack.AddrInfo
	)

	for {
		_, addr, err = s.clientConn.ReadFromUDP(receiveData)
		if err != nil {
			fmt.Println(err)
		}
		err = data.UnPackage(receiveData)
		if err != nil {
			fmt.Println(err)
		}

		if data.Type == consts.ControllerServe || data.Type == consts.RoverServe {
			fmt.Println("heartbeat received")
			err = addrInfo.UnPackage(data.Payload)
			if err != nil {
				fmt.Println(err)
			}

			if data.Type == consts.ControllerServe {
				// get the dest client from s.Groups
				s.clientMu.Lock()
				s.Groups[addrInfo.GroupId].Controller.Info = addrInfo
				s.Groups[addrInfo.GroupId].Controller.State = consts.Online
				s.clientMu.Unlock()

				s.TransMu.Lock()
				s.Groups[addrInfo.GroupId].Trans = &addrInfo.Trans
				s.TransMu.Unlock()

				// send rover addr back
				data.Type = consts.ServerResp

				data.Payload, err = s.Groups[addrInfo.GroupId].Rover.ToBytes()
				if err != nil {
					fmt.Println(err)
				}
			} else if data.Type == consts.RoverServe {
				// get the dest client from s.Groups
				s.clientMu.Lock()
				s.Groups[addrInfo.GroupId].Rover.Info = addrInfo
				s.Groups[addrInfo.GroupId].Rover.State = consts.Online
				s.clientMu.Unlock()

				// send controller addr back
				data.Type = consts.ServerResp

				data.Payload, err = s.Groups[addrInfo.GroupId].Rover.ToBytes()
				if err != nil {
					fmt.Println(err)
				}
			}

			_, err = s.clientConn.WriteTo(data.Package(), addr)
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