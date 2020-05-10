package controller

import (
	"HomeRover/base"
	"HomeRover/consts"
	gst "HomeRover/gst/gstreamer-sink"
	"HomeRover/models/config"
	"HomeRover/models/data"
	"fmt"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v2"
	"runtime"
	"strconv"
	"time"
)


func NewService(conf *config.CommonConfig, controllerConf *config.ControllerConfig, joystickData chan []byte) (service *Service, err error) {
	service = &Service{
		joystickData: 	joystickData,
	}

	service.Conf = conf
	service.controllerConf = controllerConf
	service.LocalInfo.Type = consts.Controller
	service.LocalInfo.Id = uint16(conf.Id)
	service.LocalInfo.Trans = controllerConf.Trans
	service.sdpReqSignal = make(chan bool, 1)
	return
}

type Service struct {
	base.Service

	controllerConf 	*config.ControllerConfig

	joystickData	chan []byte

	sdpReqSignal	chan bool
}

func (s *Service) cmdSend()  {
	sendObject := data.Data{
		Type:     consts.Controller,
		Channel:  consts.Cmd,
		OrderNum: 0,
		Payload:  nil,
	}

	sendEntity := data.EntityData{
		GroupId: s.DestClient.Info.GroupId,
		Payload: nil,
	}

	var (
		sendData 	[]byte
		err			error
	)

	for {
		sendEntity.Payload =  <- s.joystickData
		sendObject.Payload = sendEntity.ToBytes()
		s.DestClientMu.RLock()
		if s.DestClient.State == consts.Online {
			sendData = sendObject.ToBytes()
			_, err = s.CmdConn.WriteToUDP(sendData, s.DestClient.Info.CmdAddr)
			if err != nil {
				fmt.Println(err)
			}
		}
		s.DestClientMu.RUnlock()
	}
}

func (s *Service) cmdRecv() {
	recvBytes := make([]byte, s.Conf.PackageLen)
	recvData := data.Data{}

	for {
		_, _, err := s.CmdConn.ReadFromUDP(recvBytes)
		if err != nil {
			fmt.Println(err)
		}
		err = recvData.FromBytes(recvBytes)
		if err != nil {
			fmt.Println(err)
		}

		if recvData.Type == consts.Rover && recvData.Channel == consts.Cmd {
			fmt.Println("cmd received")
		}
	}
}

func (s *Service) videoRecv()  {
	// Prepare the configuration
	webrtcConf := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:" + s.Conf.ServerIP + ":" + strconv.Itoa(s.Conf.StunPort)},
			},
		},
	}

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(webrtcConf)
	if err != nil {
		panic(err)
	}

	// Allow us to receive 1 audio track, and 2 video tracks
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	} else if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	} else if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}

	// Set a handler for when a new remote track starts, this handler creates a gstreamer pipeline
	// for the given codec
	peerConnection.OnTrack(func(track *webrtc.Track, receiver *webrtc.RTPReceiver) {
		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		// This is a temporary fix until we implement incoming RTCP events, then we would push a PLI only when a viewer requests it
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for range ticker.C {
				rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: track.SSRC()}})
				if rtcpSendErr != nil {
					fmt.Println(rtcpSendErr)
				}
			}
		}()

		codec := track.Codec()
		fmt.Printf("Track has started, of type %d: %s \n", track.PayloadType(), codec.Name)
		pipeline := gst.CreatePipeline(codec.Name)
		pipeline.Start()
		buf := make([]byte, 1400)
		for {
			i, readErr := track.Read(buf)
			if readErr != nil {
				panic(err)
			}

			pipeline.Push(buf[:i])
		}
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(<- s.RemoteSDPCh)
	if err != nil {
		panic(err)
	}

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	s.LocalSDPCh <- answer
	go s.SendSPD(30, make(chan bool, 1))

	if len(s.WebrtcSignal) > 0 {
		<-s.WebrtcSignal
	}

	// Block forever
	select {
	case <- s.WebrtcSignal:
		runtime.Goexit()
	}
}

func (s *Service) recvSDP()  {
	recvBytes := make([]byte, s.Conf.PackageLen)
	recvData := data.Data{}
	recvSDP := data.SDPData{}
	var (
		err			error
	)

	for {
		_, _, err = s.VideoConn.ReadFromUDP(recvBytes)
		if err != nil {
			fmt.Println(err)
		}
		err = recvData.FromBytes(recvBytes)
		if err != nil {
			fmt.Println(err)
		}

		if recvData.Type == consts.Rover && recvData.Channel == consts.Video {
			err = recvSDP.FromBytes(recvData.Payload)
			if err != nil {
				fmt.Println(err)
			}
			switch recvSDP.Type {
			case consts.SDPExchange:
				s.WebrtcSignal <- true
				s.sdpReqSignal <- true
				go s.videoRecv()
			case consts.SDPReq:
			case consts.SDPEnd:
			}
		}
	}
}

func (s *Service) sendSDPReq()  {
	sendObject := data.Data{
		Type: 		consts.Controller,
		Channel: 	consts.Video,
		OrderNum: 	0,
	}

	sdp := data.SDPData{
		Type: 		consts.SDPReq,
	}

	sdpBytes, err := sdp.ToBytes()
	if err != nil {
		fmt.Println(err)
	}

	sendObject.Payload = sdpBytes
	sendData := sendObject.ToBytes()

	for range time.Tick(time.Second) {
		_, err = s.VideoConn.WriteToUDP(sendData, s.DestClient.Info.VideoAddr)
		if err != nil {
			fmt.Println(err)
		}

		select {
		case <- s.sdpReqSignal:
			return
		}
	}
}

func (s *Service) Run() {
	err := s.InitConn()
	if err != nil {
		fmt.Println(err)
	}

	go s.ServerSend()
	go s.ServerRecv()

	go s.cmdSend()
	go s.cmdRecv()

	go s.sendSDPReq()
	go s.recvSDP()

	select {}
}