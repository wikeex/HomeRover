package server

import (
	"HomeRover/consts"
	"HomeRover/log"
	"HomeRover/models/client"
	"HomeRover/models/config"
	"HomeRover/models/data"
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

	serviceAddr			*net.TCPAddr
	serviceListener 	*net.TCPListener
}

func (s *Service) init() error {
	groupSet := mapset.NewSet()
	groupSet.Add(uint16(1))
	groupSet.Add(uint16(2))
	s.Groups[0] = &server.Group{
		Id: 0,
		Rover: client.Client{Id: 1},
		Controller: client.Client{Id: 2},
	}

	var err error
	s.confMu.RLock()
	s.serviceAddr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", s.conf.ServicePort))
	if err != nil {
		return err
	}
	s.serviceListener, err = net.ListenTCP("tcp", s.serviceAddr)
	if err != nil {
		return err
	}
	s.confMu.RUnlock()

	return nil
}

func (s *Service)listenClients()  {

	log.Logger.Info("listen client task starting...")
	for {
		conn, err := s.serviceListener.Accept()
		if err != nil {
			log.Logger.Error(err)
			continue
		}
		go s.handleClient(conn)
	}
}

func (s *Service) handleClient(conn net.Conn)  {
	recvBytes := make([]byte, s.conf.PackageLen)
	recvData := data.Data{}
	var (
		err        		error
		addr       		*net.TCPAddr

		// tempClient use to resolve groupId
		tempClient		client.Client
		sourceClient	*client.Client
		destClient		*client.Client
		groupId			uint16
		onceTask		sync.Once
		length 			int
	)

	for {
		length, err = conn.Read(recvBytes)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"error": err,
			}).Error("read tcp conn error")
		}
		err = recvData.FromBytes(recvBytes[:length])
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"error": err,
			}).Error("deserialization data error")
		}

		log.Logger.WithFields(logrus.Fields{
			"received bytes": recvBytes,
			"received data": recvData,
		}).Debug("received heartbeat from client")

		err = tempClient.FromBytes(recvData.Payload)
		if err != nil {
			log.Logger.Error(err)
		}
		groupId = tempClient.GroupId

		if recvData.Type == consts.Controller {
			log.Logger.WithFields(logrus.Fields{
				"client": sourceClient,
				"addr": addr.String(),
			}).Info("controller heartbeat received")
			sourceClient = &s.Groups[groupId].Controller
			destClient = &s.Groups[groupId].Rover
		} else if recvData.Type == consts.Rover {
			log.Logger.WithFields(logrus.Fields{
				"client": sourceClient,
				"addr": addr.String(),
			}).Info("rover heartbeat received")
			sourceClient = &s.Groups[groupId].Rover
			destClient = &s.Groups[groupId].Controller
		}

		localAddr := conn.LocalAddr()
		sourceClient.Addr = &localAddr

		// get the dest client from s.Groups
		s.clientMu.Lock()
		sourceClient.State = consts.Online
		s.clientMu.Unlock()

		// create a goroutine to send sdp info
		onceTask.Do(func() {
			go func() {
				var err error
				for {
					s.clientMu.Lock()
					_, err = conn.Write(<- sourceClient.SendCh)
					if err != nil {
						log.Logger.WithFields(logrus.Fields{
							"error": err,
						}).Error("send sdp package error")
					}
					s.clientMu.Unlock()
				}
			}()
		})

		log.Logger.WithFields(logrus.Fields{
			"dest client": destClient,
		}).Debug("generating response")

		// send controller addr back
		recvData.Payload, err = destClient.ToBytes()
		if err != nil {
			log.Logger.Error(err)
		}

		if recvData.Channel == consts.Service {
			recvData.Type = consts.Server
			recvData.Channel = consts.Service

			log.Logger.WithFields(logrus.Fields{
				"data": recvData,
				"addr": addr.String(),
			}).Info("send heartbeat response to client")

			_, err = conn.Write(recvData.ToBytes())
			if err != nil {
				log.Logger.Error(err)
			}
		} else if recvData.Channel == consts.SDPReq ||
			recvData.Channel == consts.SDPExchange ||
			recvData.Channel == consts.SDPEnd {

			log.Logger.WithFields(logrus.Fields{
				"data": recvData,
				"addr": (*destClient.Addr).String(),
			}).Info("forward sdp package to destination client")

			destClient.SendCh <- recvBytes
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