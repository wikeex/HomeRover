package server

import (
	"HomeRover/consts"
	"HomeRover/log"
	"HomeRover/models/client"
	"HomeRover/models/config"
	"HomeRover/models/data"
	"HomeRover/models/mode"
	"HomeRover/models/server"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
)

func NewService(conf *config.ServerConfig) (*Service, error) {
	service := &Service{
		conf: conf,
	}
	service.Groups = make(map[uint16]*server.Group)
	return service, nil
}

type Service struct {
	conf 				*config.ServerConfig
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

	var err error
	s.confMu.RLock()
	s.serviceAddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", s.conf.ServicePort))
	if err != nil {
		return err
	}
	s.serviceConn, err = net.ListenUDP("udp", s.serviceAddr)
	if err != nil {
		return err
	}
	s.forwardAddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", s.conf.ForwardPort))
	if err != nil {
		return err
	}
	s.confMu.RUnlock()

	return nil
}

func (s *Service)listenClients()  {
	recvBytes := make([]byte, s.conf.PackageLen)
	recvData := data.Data{}
	var (
		err        		error
		addr       		*net.UDPAddr
		sourceInfo		client.Info
		sourceClient	*client.Client
		destClient		*client.Client
		groupId			uint16
	)

	log.Logger.Info("listen client task starting...")
	for {
		_, addr, err = s.serviceConn.ReadFromUDP(recvBytes)
		if err != nil {
			log.Logger.Error(err)
		}
		err = recvData.FromBytes(recvBytes)
		if err != nil {
			log.Logger.Error(err)
		}

		if recvData.Channel == consts.Service {

			err = sourceInfo.FromBytes(recvData.Payload)
			groupId = sourceInfo.GroupId
			if err != nil {
				log.Logger.Error(err)
			}

			if recvData.Type == consts.Controller {
				log.Logger.Info("controller heartbeat received")
				sourceClient = &s.Groups[groupId].Controller
				destClient = &s.Groups[groupId].Rover

				s.TransMu.Lock()
				s.Groups[groupId].Trans = &sourceInfo.Trans
				s.TransMu.Unlock()

			} else if recvData.Type == consts.Rover {
				log.Logger.Info("rover heartbeat received")
				sourceClient = &s.Groups[groupId].Rover
				destClient = &s.Groups[groupId].Controller
			}

			// get the dest client from s.Groups
			s.clientMu.Lock()
			sourceClient.Info = sourceInfo
			sourceClient.State = consts.Online
			s.clientMu.Unlock()

			// send controller addr back
			recvData.Payload, err = makeRespClientBytes(
				destClient,
				s.Groups[groupId].Trans,
				s.forwardAddr,
			)
			if err != nil {
				log.Logger.Error(err)
			}

			recvData.Type = consts.Server
			recvData.Channel = consts.Service
		}
		log.Logger.WithFields(logrus.Fields{
			"data": recvData,
			"addr": addr.String(),
		}).Info("send heartbeat response to client")
		_, err = s.serviceConn.WriteToUDP(recvData.ToBytes(), addr)
		if err != nil {
			log.Logger.Error(err)
		}
	}
}

func (s *Service) forward()  {
	recvBytes := make([]byte, s.conf.PackageLen)
	var (
		err			error
		addr		*net.UDPAddr
		recvData	data.Data
		recvEntity	data.EntityData
	)

	log.Logger.Info("forward task starting...")
	for {
		_, _, err = s.serviceConn.ReadFromUDP(recvBytes)
		if err != nil {
			log.Logger.Error(err)
		}

		err = recvData.FromBytes(recvBytes)
		if err != nil {
			log.Logger.Error(err)
		}

		err = recvEntity.FromBytes(recvData.Payload)
		if err != nil {
			log.Logger.Error(err)
		}

		log.Logger.WithFields(logrus.Fields{"data": recvEntity}).Info("forward data received")
		switch recvData.Channel {
		case consts.Cmd:
			if recvData.Type == consts.Controller {
				addr = s.Groups[recvEntity.GroupId].Rover.Info.CmdAddr
			} else {
				addr = s.Groups[recvEntity.GroupId].Controller.Info.CmdAddr
			}
		case consts.Video:
			if recvData.Type == consts.Controller {
				addr = s.Groups[recvEntity.GroupId].Rover.Info.VideoAddr
			} else {
				addr = s.Groups[recvEntity.GroupId].Controller.Info.VideoAddr
			}
		case consts.Audio:
			if recvData.Type == consts.Controller {
				addr = s.Groups[recvEntity.GroupId].Rover.Info.AudioAddr
			} else {
				addr = s.Groups[recvEntity.GroupId].Controller.Info.AudioAddr
			}
		}

		log.Logger.Info("forward data send")
		_, err = s.forwardConn.WriteToUDP(recvBytes, addr)
		if err != nil {
			log.Logger.Error(err)
		}
	}
}

func (s *Service) Run() {
	log.Logger.Info("server service starting")
	err := s.init()
	if err != nil {
		log.Logger.Error(err)
	}
	
	go s.listenClients()
	select {}
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