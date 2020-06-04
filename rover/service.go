package rover

import (
	"HomeRover/base"
	"HomeRover/consts"
	gst "HomeRover/gst/gstreamer-src"
	"HomeRover/log"
	"HomeRover/models/config"
	"HomeRover/utils"
	"bytes"
	"flag"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v2"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"time"
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

	audioSrc		*string
	videoSrc		*string
	webrtcConf		webrtc.Configuration
}

func (s *Service) initWebrtc()  {
	s.audioSrc = flag.String("audio-src", "audiotestsrc", "GStreamer audio src")
	s.videoSrc = flag.String("video-src", "v4l2src ! image/jpeg,width=1280,height=960,framerate=30/1 ! jpegparse ! jpegdec", "GStreamer video src")
	flag.Parse()

	// Prepare the configuration
	s.webrtcConf = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:" + s.Conf.ServerIP + ":" + strconv.Itoa(s.Conf.StunPort)},
			},
		},
	}
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

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(s.webrtcConf)
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

	var answer webrtc.SessionDescription
	select {
	case answer = <- s.RemoteSDPCh:
		log.Logger.WithFields(logrus.Fields{
			"remote sdp": utils.Encode(answer),
		}).Debug("got remote sdp from remote sdp channel")
	case <- s.WebrtcEndSignal:
		log.Logger.Info("webrtc exit signal got, restart webrtc")
		err = peerConnection.Close()
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("close peerConnection err")
		}
		runtime.Goexit()
	}


	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(answer)
	if err != nil {
		panic(err)
	}

	// Start pushing buffers on these tracks
	audioPipeline := gst.CreatePipeline(webrtc.Opus, []*webrtc.Track{audioTrack}, *s.audioSrc)
	videoPipeline := gst.CreatePipeline(webrtc.H264, []*webrtc.Track{videoTrack}, *s.videoSrc)

	audioPipeline.Start()
	videoPipeline.Start()
	// Block forever
	select {
	case <- s.WebrtcEndSignal:
		log.Logger.Info("webrtc exit signal got, restart webrtc")
		audioPipeline.Stop()
		videoPipeline.Stop()
		err = peerConnection.Close()
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("close peerConnection err")
		}
		runtime.Goexit()
	}
}

func (s *Service) webrtcGstreamerCli()  {
	log.Logger.Info("webrtc task starting...")

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


	// Open a UDP Listener for RTP Packets on port 5004
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5004})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = listener.Close(); err != nil {
			panic(err)
		}
	}()

	log.Logger.Info("Waiting for RTP Packets, please run GStreamer or ffmpeg now")

	// Listen for a single RTP Packet, we need this to determine the SSRC
	inboundRTPPacket := make([]byte, 4096) // UDP MTU
	n, _, err := listener.ReadFromUDP(inboundRTPPacket)
	if err != nil {
		panic(err)
	}

	// Unmarshal the incoming packet
	packet := &rtp.Packet{}
	if err = packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
		panic(err)
	}

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Logger.Info("Connection State has changed %s \n", connectionState.String())
	})

	// Create a video track
	videoTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeH264, packet.SSRC, "video", "pion2")
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

	for {
		n, _, err := listener.ReadFrom(inboundRTPPacket)
		if err != nil {
			log.Logger.Error("error during read: %s", err)
			panic(err)
		}

		packet := &rtp.Packet{}
		if err := packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
			panic(err)
		}
		packet.Header.PayloadType = webrtc.DefaultPayloadTypeH264

		if writeErr := videoTrack.WriteRTP(packet); writeErr != nil {
			panic(writeErr)
		}

		select {
		case <- s.WebrtcEndSignal:
			log.Logger.Info("webrtc exit signal got, restart webrtc")
			err = peerConnection.Close()
			if err != nil {
				log.Logger.WithFields(logrus.Fields{
					"err": err,
				}).Error("close peerConnection err")
			}
			runtime.Goexit()
		default:
		}
	}
}

func (s *Service) startGstreamer()  {
	var err		error
	var stdout 	bytes.Buffer
	var stderr 	bytes.Buffer

	// start gstreamer v4l2 video
	cmd := exec.Command( //nolint
		"gst-launch-1.0",
		"-v",
		"v4l2src",
		"!", `videoconvert`,
		"!", "omxh264enc",
		"!", `h264parse`,
		"!", "rtph264pay", "config-interval=10", "pt=96",
		"!", "udpsink", "host=127.0.0.1", "port=5004", "-e",
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	log.Logger.WithFields(logrus.Fields{
		"stdout": cmd.Stdout,
		"stderr": cmd.Stderr,
	}).Info("execute gst command")
	if err = cmd.Run(); err != nil {
		log.Logger.WithFields(logrus.Fields{
			"stdout": cmd.Stdout,
			"stderr": cmd.Stderr,
		}).Error("execute gst command occur an error")
	}
}


func (s *Service) Run()  {
	log.Logger.Info("rover service starting")
	err := s.InitConn()
	if err != nil {
		log.Logger.Error(err)
	}

	s.initWebrtc()

	go s.Send()
	go s.ServerRecv()

	go s.SignIn()
	go s.cmdService()

	if s.Conf.GstreamerCli {
		go s.startGstreamer()
	}

	for {
		if s.Conf.GstreamerCli {
			go s.webrtcGstreamerCli()
		} else {
			go s.webrtc()
		}
		select {
		case <- s.SDPReqCh:
			s.WebrtcEndSignal <- true
			time.Sleep(time.Second)
		}
	}

}