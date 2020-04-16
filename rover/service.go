package rover

import (
	"HomeRover/base"
	"HomeRover/consts"
	"HomeRover/models/config"
)

func NewService(conf *config.CommonConfig, roverConf *config.RoverConfig, joystickData chan []byte) (service *Service, err error) {
	service = &Service{
		joystickData: 	joystickData,
	}

	service.Conf = conf
	service.roverConf = roverConf
	service.LocalInfo.Type = consts.Rover
	service.LocalInfo.Id = uint16(conf.Id)
	return
}

type Service struct {
	base.Service

	roverConf 		*config.RoverConfig
	joystickData	chan []byte
}


