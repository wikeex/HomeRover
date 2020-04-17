package rover

import (
	"HomeRover/base"
	"HomeRover/consts"
	"HomeRover/models/config"
	"HomeRover/models/data"
	"fmt"
	"net"
)

func NewService(conf *config.CommonConfig, roverConf *config.RoverConfig, joystickDataCh chan []byte) (service *Service, err error) {
	service = &Service{
		joystickDataCh: 	joystickDataCh,
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
	joystickDataCh	chan []byte

	cmdServiceConn	*net.UDPConn
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
			
			s.joystickDataCh <- recvEntity.Payload
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

func (s *Service) cmdService()  {
	cmdServiceAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", s.roverConf.CmdServicePort))
	if err != nil {
		fmt.Println(err)
	}
	sendAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	s.cmdServiceConn, err = net.ListenUDP("udp", sendAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer s.cmdServiceConn.Close()

	for {
		_, err = s.cmdServiceConn.WriteToUDP(<- s.joystickDataCh, cmdServiceAddr)
		if err != nil {
			fmt.Println(err)
		}
	}
}