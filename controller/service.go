package controller

import (
	"HomeRover/base"
	"HomeRover/consts"
	gst "HomeRover/gst/gstreamer-sink"
	"HomeRover/log"
	"HomeRover/models/config"
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
	service.LocalClient.Type = consts.Controller
	service.LocalClient.Id = uint16(conf.Id)
	service.sdpReqSignal = make(chan bool, 1)
	return
}

type Service struct {
	base.Service

	controllerConf 	*config.ControllerConfig

	joystickData	chan []byte

	sdpReqSignal	chan bool
}

func (s *Service) webrtc()  {
	log.Logger.Info("webrtc task starting...")

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

	// Allow us to receive 1 audio track and 1 video track, send 1 command data channel and 1 audio track
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
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
					log.Logger.Error(rtcpSendErr)
				}
			}
		}()

		codec := track.Codec()
		log.Logger.Info("Track has started, of type %d: %s \n", track.PayloadType(), codec.Name)
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
		log.Logger.Info("Connection State has changed %s \n", connectionState.String())
	})

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		log.Logger.Info("New DataChannel %s %d\n", d.Label(), d.ID())

		// Register channel opening handling
		d.OnOpen(func() {
			log.Logger.Info("Data channel '%s'-'%d' open. Joystick data will now be sent to any connected DataChannels\n", d.Label(), d.ID())

			for {
				// Send the message as bytes
				sendErr := d.Send(<- s.joystickData)
				if sendErr != nil {
					panic(sendErr)
				}
			}
		})

		// Register text message handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Logger.Info("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
		})
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

	if len(s.WebrtcSignal) > 0 {
		<-s.WebrtcSignal
	}

	// Block forever
	select {
	case <- s.WebrtcSignal:
		log.Logger.Info("got exit webrtc signal, webrtc will exit")
		runtime.Goexit()
	}
}

func (s *Service) Run() {
	log.Logger.Info("controller service starting...")
	err := s.InitConn()
	if err != nil {
		log.Logger.Error(err)
	}

	go s.ServerSend()
	go s.ServerRecv()

	go s.webrtc()
	select {}
}