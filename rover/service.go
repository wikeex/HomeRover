package rover

import (
	"HomeRover/base"
	"HomeRover/consts"
	gst "HomeRover/gst/gstreamer-src"
	"HomeRover/log"
	"HomeRover/models/config"
	"HomeRover/utils"
	"flag"
	"github.com/pion/webrtc/v2"
	"github.com/sirupsen/logrus"
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
	service.LocalClient.Type = consts.Rover
	service.LocalClient.Id = uint16(conf.Id)
	return
}

type Service struct {
	base.Service

	roverConf 		*config.RoverConfig
	joystickDataCh	chan []byte

	cmdServiceConn	net.Conn
}

func (s *Service) cmdService()  {
	var err error
	s.cmdServiceConn, err = net.Dial("udp", "127.0.0.1:" + strconv.Itoa(s.roverConf.CmdServicePort))
	if err != nil {
		log.Logger.Error(err)
	}
	defer func() {
		err := s.cmdServiceConn.Close()
		if err != nil {
			log.Logger.Error(err)
		}
	}()

	log.Logger.Info("command to device driver task start...")
	for {
		_, err = s.cmdServiceConn.Write(<- s.joystickDataCh)
		if err != nil {
			log.Logger.Error(err)
		}
	}
}

func (s *Service) webrtc()  {
	log.Logger.Info("webrtc task starting...")

	audioSrc := flag.String("audio-src", "audiotestsrc", "GStreamer audio src")
	videoSrc := flag.String("video-src", "v4l2src ! videoconvert", "GStreamer video src")
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


	// Create a datachannel with label 'data'
	dataChannel, err := peerConnection.CreateDataChannel("cmdData", nil)
	if err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Logger.Info("Connection State has changed %s \n", connectionState.String())
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
	videoTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeH264, rand.Uint32(), "video", "pion2")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		panic(err)
	}

	// Register text message handling
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		s.joystickDataCh <- msg.Data
	})

	// create offer from peer
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		panic(err)
	}

	s.SendCh <- utils.Encode(offer)

	var answer = <- s.RemoteSDPCh
	log.Logger.WithFields(logrus.Fields{
		"remote sdp": utils.Encode(answer),
	}).Debug("got remote sdp from remote sdp channel")

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(answer)
	if err != nil {
		panic(err)
	}

	// Start pushing buffers on these tracks
	gst.CreatePipeline(webrtc.Opus, []*webrtc.Track{audioTrack}, *audioSrc).Start()
	gst.CreatePipeline(webrtc.H264, []*webrtc.Track{videoTrack}, *videoSrc).Start()

	// Block forever
	select {
	case <- s.WebrtcSignal:
		log.Logger.Info("video send task end")
		runtime.Goexit()
	}
}

func (s *Service) Run()  {
	log.Logger.Info("rover service starting")
	err := s.InitConn()
	if err != nil {
		log.Logger.Error(err)
	}

	go s.Send()
	go s.ServerRecv()

	go s.SignIn()
	go s.webrtc()
	go s.cmdService()

	select {}
}