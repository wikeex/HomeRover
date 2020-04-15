package server

import (
	"HomeRover/models/client"
	"HomeRover/models/config"
	"HomeRover/models/consts"
	"HomeRover/models/data"
	"HomeRover/models/mode"
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

	serviceAddr			*net.UDPAddr
	forwardAddr			*net.UDPAddr

	serviceConn 		*net.UDPConn
	forwardConn			*net.UDPConn
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

	s.serviceConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	
	return nil
}

func (s *Service)listenClients()  {
	recvBytes := make([]byte, s.conf.PackageLen)
	recvData := data.Data{}
	var (
		err        error
		addr       *net.UDPAddr
		clientInfo client.Info
	)

	for {
		_, addr, err = s.serviceConn.ReadFromUDP(recvBytes)
		if err != nil {
			fmt.Println(err)
		}
		err = recvData.FromBytes(recvBytes)
		if err != nil {
			fmt.Println(err)
		}

		if recvData.Type == consts.ControllerServe || recvData.Type == consts.RoverServe {
			fmt.Println("heartbeat received")
			err = clientInfo.FromBytes(recvData.Payload)
			if err != nil {
				fmt.Println(err)
			}

			if recvData.Type == consts.ControllerServe {
				// get the dest client from s.Groups
				s.clientMu.Lock()
				s.Groups[clientInfo.GroupId].Controller.Info = clientInfo
				s.Groups[clientInfo.GroupId].Controller.State = consts.Online
				s.clientMu.Unlock()

				s.TransMu.Lock()
				s.Groups[clientInfo.GroupId].Trans = &clientInfo.Trans
				s.TransMu.Unlock()

				// send rover addr back
				recvData.Type = consts.ServerResp

				recvData.Payload, err = makeRespClientBytes(
					&s.Groups[clientInfo.GroupId].Rover,
					s.Groups[clientInfo.GroupId].Trans,
					s.forwardAddr,
				)
				if err != nil {
					fmt.Println(err)
				}
			} else if recvData.Type == consts.RoverServe {
				// get the dest client from s.Groups
				s.clientMu.Lock()
				s.Groups[clientInfo.GroupId].Rover.Info = clientInfo
				s.Groups[clientInfo.GroupId].Rover.State = consts.Online
				s.clientMu.Unlock()

				// send controller addr back
				recvData.Type = consts.ServerResp

				recvData.Payload, err = makeRespClientBytes(
					&s.Groups[clientInfo.GroupId].Controller,
					s.Groups[clientInfo.GroupId].Trans,
					s.forwardAddr,
				)
				if err != nil {
					fmt.Println(err)
				}
			}

			_, err = s.serviceConn.WriteToUDP(recvData.ToBytes(), addr)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (s *Service) forward()  {
	recvBytes := make([]byte, s.conf.PackageLen)
	recvData := data.Data{}
	var (
		err			error
		addr		*net.UDPAddr
		sendBytes	[]byte
	)
	for {
		_, _, err = s.serviceConn.ReadFromUDP(recvBytes)
		if err != nil {
			fmt.Println(err)
		}

		err = recvData.FromBytes(recvBytes)
		if err != nil {
			fmt.Println(err)
		}

		switch recvData.Type {
		case consts.ControllerCmd:
		case consts.ControllerVideo:
		case consts.ControllerAudio:
		case consts.RoverCmd:
		case consts.RoverVideo:
		case consts.RoverAudio:
		}

		sendBytes = recvData.ToBytes()

		_, err = s.forwardConn.WriteToUDP(sendBytes, addr)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (s *Service) Run() {
	err := s.init()
	if err != nil {
		fmt.Println(err)
	}
	
	go s.listenClients()
}

func makeRespClientBytes(c *client.Client, transRule *mode.Trans, forwardAddr *net.UDPAddr) ([]byte, error) {
	respClient := client.Client{
		State: c.State,
		Info:  client.Info{},
	}

	if transRule.Cmd {
		// if cmd channel is HoldPunching mode
		respClient.Info.CmdAddr = c.Info.CmdAddr
	} else {
		// cmd channel is forwarding mode
		respClient.Info.CmdAddr = forwardAddr
	}

	if transRule.Video {
		respClient.Info.VideoAddr = c.Info.VideoAddr
	} else {
		respClient.Info.VideoAddr = forwardAddr
	}

	if transRule.Audio {
		respClient.Info.AudioAddr = c.Info.AudioAddr
	} else {
		respClient.Info.AudioAddr = forwardAddr
	}

	respBytes, err := respClient.ToBytes()
	if err != nil {
		return nil, err
	}

	return respBytes, nil
}