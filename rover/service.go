package rover

import (
	"HomeRover/base"
	"HomeRover/consts"
	gst "HomeRover/gst/gstreamer-src"
	"HomeRover/models/config"
	"HomeRover/models/data"
	"flag"
	"fmt"
	"github.com/pion/webrtc/v2"
	"math/rand"
	"net"
	"runtime"
	"strconv"
)

func NewService(conf *config.CommonConfig, roverConf *config.RoverConfig) (service *Service, err error) {
	service = &Service{}

	service.Conf = conf
	service.roverConf = roverConf
	service.joystickDataCh = make(chan []byte, 1)
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
		err			error
	)

	for {
		_, _, err = s.CmdConn.ReadFromUDP(recvBytes)
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

	sendAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:" + strconv.Itoa(s.roverConf.CmdServicePort))
	s.cmdServiceConn, err = net.ListenUDP("udp", sendAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		err := s.cmdServiceConn.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	for {
		_, err = s.cmdServiceConn.WriteToUDP(<- s.joystickDataCh, cmdServiceAddr)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (s *Service) webrtc()  {
	audioSrc := flag.String("audio-src", "audiotestsrc", "GStreamer audio src")
	videoSrc := flag.String("video-src", "v4l2src ! 'video/x-raw,width=1280, height=960, framerate=30/1' ! nvvidconv ", "GStreamer video src")
	flag.Parse()

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

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
	})

	// Create a audio track
	audioTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "audio", "pion1")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		panic(err)
	}

	// Create a video track
	firstVideoTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeH264, rand.Uint32(), "video", "pion2")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(firstVideoTrack)
	if err != nil {
		panic(err)
	}

	// Create a second video track
	secondVideoTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeH264, rand.Uint32(), "video", "pion3")
	if err != nil {
		panic(err)
	}

	_, err = peerConnection.AddTrack(secondVideoTrack)
	if err != nil {
		panic(err)
	}

	// create offer from peer
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		panic(err)
	}

	s.LocalSDPCh <- offer
	end := make(chan bool, 1)
	go s.SendSPD(0, end)

	// When received answer from remote, end the SDP send goroutine
	answer := <- s.RemoteSDPCh
	end <- true

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(answer)
	if err != nil {
		panic(err)
	}

	// Start pushing buffers on these tracks
	gst.CreatePipeline(webrtc.Opus, []*webrtc.Track{audioTrack}, *audioSrc).Start()
	gst.CreatePipeline(webrtc.H264, []*webrtc.Track{firstVideoTrack, secondVideoTrack}, *videoSrc).Start()

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
				s.RemoteSDPCh <- recvSDP.SDPInfo
			case consts.SDPReq:
				s.WebrtcSignal <- true
				go s.webrtc()
			case consts.SDPEnd:
			}
		}
	}
}

func (s *Service) Run()  {
	err := s.InitConn()
	if err != nil {
		fmt.Println(err)
	}

	go s.ServerSend()
	go s.ServerRecv()

	go s.cmdRecv()
	go s.cmdService()

	go s.SendSPD(0, s.WebrtcSignal)
	go s.recvSDP()

	select {}
}