package base

import (
	"HomeRover/log"
	"HomeRover/models/client"
	"HomeRover/models/config"
	"HomeRover/utils"
	"encoding/json"
	"fmt"
	"github.com/fwhezfwhez/tcpx"
	"github.com/pion/webrtc/v2"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"strconv"
	"sync"
)

type Service struct {
	Conf 			*config.CommonConfig

	ServerConn 		net.Conn

	LocalClient		client.Client
	LocalClientMu 	sync.RWMutex

	DestClient		client.Client
	DestClientMu	sync.RWMutex

	RemoteSDPCh		chan webrtc.SessionDescription
	SendCh			chan string
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
	s.SendCh = make(chan string, 1)
	s.WebrtcSignal = make(chan bool, 1)

	log.Logger.WithFields(logrus.Fields{
		"server port": s.ServerConn.LocalAddr().String(),
	}).Info("allocated server port")

	return nil
}

func (s *Service) Send() {
	var  (
		buf 	[]byte
		err		error
	)

	for {
		buf, err = tcpx.PackWithMarshaller(tcpx.Message{
			MessageID: 5,
			Header:    nil,
			Body: struct {
				Message string `json:"message"`
				ToUser  string `json:"to_user"`
			}{Message: <-s.SendCh, ToUser: strconv.Itoa(s.Conf.RemoteId)},
		}, tcpx.JsonMarshaller{})
		if err != nil {
			panic(err)
		}

		_, err = s.ServerConn.Write(buf)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"error": err,
			}).Error("send data error")
		}

		log.Logger.Debug("get data from send channel and sent succeed")
	}
}

func (s *Service) ServerRecv()  {
	var buf = make([]byte, 20480)
	var e error
	for {
		buf, e = tcpx.FirstBlockOf(s.ServerConn)
		if e != nil {
			fmt.Println(e.Error())
			os.Exit(0)
		}
		bf, _ := tcpx.BodyBytesOf(buf)
		messageID, e := tcpx.MessageIDOf(buf)

		switch messageID {
		case 400, 200, 500:
			type ResponseTo struct {
				Message string `json:"message"`
			}
			var rs ResponseTo
			e = json.Unmarshal(bf, &rs)
			if e != nil {
				panic(e)
			}

			log.Logger.WithFields(logrus.Fields{
				"200 server Message": rs.Message,
			}).Info("got server message")

		case 6:
			type ResponseTo struct {
				Message  string `json:"message"`
				FromUser string `json:"from_user"`
			}
			var rs ResponseTo
			e = json.Unmarshal(bf, &rs)
			if e != nil {
				panic(e)
			}

			log.Logger.WithFields(logrus.Fields{
				"remote Message": rs.Message,
			}).Info("got remote message")

			remoteOffer := webrtc.SessionDescription{}
			utils.Decode(rs.Message, &remoteOffer)

			s.RemoteSDPCh <- remoteOffer
		}
	}
}

func (s *Service) SignIn() {
	onlineBuf, e := tcpx.PackWithMarshaller(tcpx.Message{
		MessageID: 1,
		Header:    nil,
		Body:      struct{ Username string `json:"username"` }{strconv.Itoa(s.Conf.Id)},
	}, tcpx.JsonMarshaller{})
	if e != nil {
		panic(e)
	}
	_, err := s.ServerConn.Write(onlineBuf)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("sign in server error")
	}
	log.Logger.WithFields(logrus.Fields{
		"data": onlineBuf,
	}).Info("sign in success")
}