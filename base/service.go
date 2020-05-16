package base

import (
	"HomeRover/consts"
	"HomeRover/log"
	"HomeRover/models/client"
	"HomeRover/models/config"
	"HomeRover/models/data"
	"fmt"
	"github.com/pion/webrtc/v2"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

func allocatePort(conn **net.UDPConn) (*net.UDPAddr, error) {
	rand.Seed(time.Now().UnixNano())
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", rand.Intn(55535) + 10000))
	if err != nil {
		return allocatePort(conn)
	}
	*conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	return addr, nil
}


type Service struct {
	Conf 			*config.CommonConfig

	ServerConn 		*net.UDPConn
	CmdConn    		*net.UDPConn
	VideoConn  		*net.UDPConn
	AudioConn  		*net.UDPConn

	DestClient     	client.Client
	DestClientMu   	sync.RWMutex

	LocalInfo 		client.Info
	LocalInfoMu    	sync.RWMutex

	RemoteSDPCh		chan webrtc.SessionDescription
	LocalSDPCh		chan webrtc.SessionDescription
	WebrtcSignal	chan bool
}

func (s *Service) InitConn() error {
	_, err := allocatePort(&s.ServerConn)

	s.LocalInfoMu.Lock()
	if err != nil {
		return err
	}
	s.LocalInfo.CmdAddr, err = allocatePort(&s.CmdConn)
	if err != nil {
		return err
	}
	s.LocalInfo.VideoAddr, err = allocatePort(&s.VideoConn)
	if err != nil {
		return err
	}
	s.LocalInfo.AudioAddr, err = allocatePort(&s.AudioConn)
	if err != nil {
		return err
	}
	s.LocalInfoMu.Unlock()

	s.RemoteSDPCh = make(chan webrtc.SessionDescription, 1)
	s.LocalSDPCh = make(chan webrtc.SessionDescription, 1)
	s.WebrtcSignal = make(chan bool, 1)

	log.Logger.WithFields(logrus.Fields{
		"cmd addr": s.LocalInfo.CmdAddr.String(),
		"video addr": s.LocalInfo.VideoAddr.String(),
		"audio addr": s.LocalInfo.AudioAddr.String(),
	}).Info("allocated local port")

	return nil
}

func (s *Service) ServerSend()  {
	s.LocalInfoMu.RLock()
	addrBytes, err := s.LocalInfo.ToBytes()
	if err != nil {
		log.Logger.Error(err)
	}
	sendObject := data.Data{
		Type:     s.LocalInfo.Type,
		Channel:  consts.Service,
		OrderNum: 0,
		Payload:  addrBytes,
	}
	s.LocalInfoMu.RUnlock()

	sendData := sendObject.ToBytes()

	addrStr := s.Conf.ServerIP + ":" + strconv.Itoa(s.Conf.ServerPort)
	addr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		log.Logger.Error(err)
	}

	log.Logger.Info("starting server send task")
	for range time.Tick(time.Second){
		log.Logger.WithFields(logrus.Fields{
			"data": sendData,
			"addr": addr.String(),
		}).Debug("send heartbeat to server")
		_, err = s.ServerConn.WriteToUDP(sendData, addr)
		if err != nil {
			log.Logger.Error(err)
		}
		sendObject.OrderNum++
		sendData = sendObject.ToBytes()
	}
}

func (s *Service) ServerRecv()  {
	receiveData := make([]byte, s.Conf.PackageLen)
	RecvData := data.Data{}

	for {
		_, _, err := s.ServerConn.ReadFromUDP(receiveData)
		if err != nil {
			log.Logger.Error(err)
		}
		err = RecvData.FromBytes(receiveData)
		if err != nil {
			log.Logger.Error(err)
		}

		log.Logger.WithFields(logrus.Fields{

		})

		log.Logger.Info("starting server receive task")
		if RecvData.Type == consts.Server && RecvData.Channel == consts.Service {
			s.DestClientMu.Lock()
			err = s.DestClient.FromBytes(RecvData.Payload)
			if err != nil {
				log.Logger.Error(err)
			}
			s.DestClientMu.Unlock()
			log.Logger.WithFields(logrus.Fields{
				"destination client": s.DestClient,
			}).Info("received heartbeat response")
			s.DestClientMu.RLock()
			if s.DestClient.State == consts.Offline {
				fmt.Println("rover is offline")
			}
			s.DestClientMu.RUnlock()
		}
	}
}

// Video Channel use to send spd now
func (s *Service) SendSDP(second uint16, endSignal chan bool)  {
	s.LocalInfoMu.RLock()
	sendObject := data.Data{
		Type: 		s.LocalInfo.Type,
		Channel: 	consts.Video,
		OrderNum: 	0,
	}
	s.LocalInfoMu.RUnlock()

	var (
		sdp			data.SDPData
		err			error
		timeout	= make(chan bool, 1)
	)

	if second > 0 {
		go func() {
			time.Sleep(time.Duration(second) * time.Second)
			timeout <- true
		}()
	}

	sdp.Type = consts.SDPExchange
	sdp.SDPInfo = <- s.LocalSDPCh

	sendObject.Payload, err = sdp.ToBytes()
	if err != nil {
		log.Logger.Error(err)
	}

	log.Logger.Info("start sdp task")
	for range time.Tick(3000 * time.Millisecond){
		s.LocalInfoMu.RLock()
		_, err = s.VideoConn.WriteToUDP(sendObject.ToBytes(), s.LocalInfo.VideoAddr)
		s.LocalInfoMu.RUnlock()
		log.Logger.WithFields(logrus.Fields{
			"sdp info": sdp.SDPInfo,
		}).Info("sdp info sent")

		if err != nil {
			log.Logger.Error(err)
		}

		select {
		case <- timeout:
			return
		case <- endSignal:
			return
		default:
		}
	}
}