package server

import (
	"HomeRover/log"
	"HomeRover/models/client"
	"HomeRover/models/config"
	"HomeRover/models/server"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/fwhezfwhez/errorx"
	"github.com/fwhezfwhez/tcpx"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
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
	clientMu			sync.RWMutex

	addr 				string
	service 			*tcpx.TcpX
}

func (s *Service) init() error {
	groupSet := mapset.NewSet()
	groupSet.Add(uint16(1))
	groupSet.Add(uint16(2))
	s.Groups[0] = &server.Group{
		Id: 0,
		Rover: client.Client{Id: 1, SendCh: make(chan []byte, 1)},
		Controller: client.Client{Id: 2, SendCh: make(chan []byte, 1)},
	}

	s.addr = fmt.Sprintf("0.0.0.0:%d", s.conf.ServicePort)
	s.service = tcpx.NewTcpX(tcpx.JsonMarshaller{})
	s.service.HeartBeatModeDetail(true, 3*time.Second, false, tcpx.DEFAULT_HEARTBEAT_MESSAGEID)
	s.service.WithBuiltInPool(true)

	s.service.AddHandler(1, s.online)
	s.service.AddHandler(3, s.offline)
	s.service.AddHandler(5, s.send)

	return nil
}

func (s *Service) online(c *tcpx.Context) {
	type Login struct {
		Username string `json:"username"`
	}
	var login Login
	if _, e := c.Bind(&login); e != nil {
		log.Logger.WithFields(logrus.Fields{
			"error": errorx.Wrap(e).Error(),
		}).Error("bind online function error")
		return
	}
	err := c.Online(login.Username)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"error": errorx.Wrap(err).Error(),
		}).Error("set client online error")
	}
}

func (s *Service) offline(c *tcpx.Context) {
	err :=c.Offline()
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"error": errorx.Wrap(err).Error(),
		}).Error("set client offline error")
	}
}

func (s *Service) send(c *tcpx.Context) {
	type RequestFrom struct {
		Message string `json:"message"`
		ToUser  string `json:"to_user"`
	}
	type ResponseTo struct {
		Message  string `json:"message"`
		FromUser string `json:"from_user"`
	}
	var req RequestFrom
	if _, e := c.Bind(&req); e != nil {
		panic(e)
	}
	anotherCtx := c.GetPoolRef().GetClientPool(req.ToUser)
	if anotherCtx.IsOnline() {
		err := c.SendToUsername(req.ToUser, 6, ResponseTo{
			Message:  req.Message,
			FromUser: req.ToUser,
		})
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"error": errorx.Wrap(err).Error(),
			}).Error("forward message to dest client error")
		}
	} else {
		err := c.JSON(200, ResponseTo{
			Message: fmt.Sprintf("'%s' is offline", req.ToUser),
		})
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"error": errorx.Wrap(err).Error(),
			}).Error("response to client error")
		}
	}
}

func (s *Service) Run() {
	log.Logger.Info("server service starting")
	err := s.init()
	if err != nil {
		log.Logger.Error(err)
	}

	go func() {
		err := s.service.ListenAndServe("tcp", s.addr)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"error": err,
			}).Error("start tcpx server error")
		}
	}()

	select {}
}