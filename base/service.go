package base

import (
	"HomeRover/consts"
	"HomeRover/log"
	"HomeRover/models/client"
	"HomeRover/models/config"
	"HomeRover/models/data"
	"HomeRover/utils"
	"github.com/pion/webrtc/v2"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
	"sync"
	"time"
)

type Service struct {
	Conf 			*config.CommonConfig

	ServerConn 		net.Conn

	LocalClient		client.Client
	LocalClientMu 	sync.RWMutex

	DestClient		client.Client
	DestClientMu	sync.RWMutex

	RemoteSDPCh		chan webrtc.SessionDescription
	LocalSDPCh		chan webrtc.SessionDescription
	WebrtcSignal	chan bool
}

func (s *Service) InitConn() error {
	addrStr := s.Conf.ServerIP + ":" + strconv.Itoa(s.Conf.ServerPort)
	addr, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("resolve server addr error")
	}
	s.ServerConn, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("dial server error")
		panic(err)
	}

	s.RemoteSDPCh = make(chan webrtc.SessionDescription, 1)
	s.LocalSDPCh = make(chan webrtc.SessionDescription, 1)
	s.WebrtcSignal = make(chan bool, 1)

	log.Logger.WithFields(logrus.Fields{
		"server port": s.ServerConn.LocalAddr().String(),
	}).Info("allocated server port")

	return nil
}

func (s *Service) ServerSend()  {
	s.LocalClientMu.RLock()
	clientBytes, err := s.LocalClient.ToBytes()
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"local client": s.LocalClient,
		}).Error(err)
	}
	sendObject := data.Data{
		Type:     s.LocalClient.Type,
		Channel:  consts.Service,
		OrderNum: 0,
		Payload:  clientBytes,
	}
	s.LocalClientMu.RUnlock()

	sendData := sendObject.ToBytes()

	log.Logger.Info("starting server send task")
	for range time.Tick(time.Second){
		log.Logger.WithFields(logrus.Fields{
			"info data": s.LocalClient,
			"send bytes": sendData,
			"addr": s.ServerConn.LocalAddr().String(),
		}).Debug("send heartbeat to server")

		_, err = s.ServerConn.Write(sendData)
		if err != nil {
			log.Logger.Error(err)
		}

		select {
		case localOffer := <- s.LocalSDPCh:
			sendObject.Payload = []byte(utils.Encode(localOffer))
			sendObject.Channel = consts.SDPExchange
		default:
			sendObject.OrderNum++
		}
		sendData = sendObject.ToBytes()
	}
}

func (s *Service) ServerRecv()  {
	recvBytes := make([]byte, s.Conf.PackageLen)
	recvData := data.Data{}

	log.Logger.Info("starting server receive task")
	for {
		length, err := s.ServerConn.Read(recvBytes)
		if err != nil {
			log.Logger.Error(err)
		}
		err = recvData.FromBytes(recvBytes[:length])
		if err != nil {
			log.Logger.Error(err)
		}

		log.Logger.WithFields(logrus.Fields{
			"response data": recvData,
		}).Info("received heartbeat response")

		if recvData.Type == consts.Server && recvData.Channel == consts.Service {
			s.DestClientMu.Lock()
			err = s.DestClient.FromBytes(recvData.Payload)
			if err != nil {
				log.Logger.Error(err)
			}
			s.DestClientMu.Unlock()

			s.DestClientMu.RLock()
			if s.DestClient.State == consts.Offline {
				log.Logger.Info("rover is offline")
			}
			s.DestClientMu.RUnlock()
		} else if recvData.Channel == consts.SDPExchange {
			remoteOffer := webrtc.SessionDescription{}
			utils.Decode(string(recvData.Payload), &remoteOffer)
			s.RemoteSDPCh <- remoteOffer
		} else if recvData.Channel == consts.SDPReq {
			s.WebrtcSignal <- true
		}
	}
}
