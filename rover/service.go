package rover

import (
	"HomeRover/base"
	"HomeRover/consts"
	"HomeRover/models/config"
	"HomeRover/models/data"
	"fmt"
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

func (s *Service) cmdRecv()  {
	recvBytes := make([]byte, s.Conf.PackageLen)
	recvData := data.Data{}
	recvEntity := data.EntityData{}
	var (
		counter 	uint8
		sendData 	data.Data
		sendEntity	data.EntityData
	)

	for {
		_, _, err := s.CmdConn.ReadFromUDP(recvBytes)
		if err != nil {
			fmt.Println(err)
		}
		err = recvData.FromBytes(recvBytes)
		if err != nil {
			fmt.Println(err)
		}

		if recvData.Type == consts.Controller && recvData.Channel == consts.Cmd {
			fmt.Println("cmd received")
			err = recvEntity.FromBytes(recvData.Payload)
			if err != nil {
				fmt.Println(err)
			}
			
			s.joystickData <- recvEntity.Payload
			counter++
			if counter == 255 {
				sendEntity.GroupId = recvEntity.GroupId
				sendData.Payload = sendEntity.ToBytes()
				sendData.Type = consts.Rover
				sendData.Channel = consts.Cmd
				_, err = s.CmdConn.WriteToUDP(sendData.ToBytes(), s.DestClient.Info.CmdAddr)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}

func cmdService()  {
	
}