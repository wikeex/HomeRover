package controller

import (
	"HomeRover/base"
	"HomeRover/consts"
	"HomeRover/log"
	"HomeRover/models/config"
	"HomeRover/utils"
	"bytes"
	"context"
	"fmt"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v2"
	"github.com/sirupsen/logrus"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

type udpConn struct {
	conn *net.UDPConn
	port int
}

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

	// Create context
	ctx, cancel := context.WithCancel(context.Background())

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

	// Create a local addr
	var laddr *net.UDPAddr
	if laddr, err = net.ResolveUDPAddr("udp", "127.0.0.1:"); err != nil {
		panic(err)
	}

	// Prepare udp conns
	udpConns := map[string]*udpConn{
		"audio": {port: 4000},
		"video": {port: 5004},
	}
	for _, c := range udpConns {
		// Create remote addr
		var raddr *net.UDPAddr
		if raddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", c.port)); err != nil {
			panic(err)
		}

		// Dial udp
		if c.conn, err = net.DialUDP("udp", laddr, raddr); err != nil {
			panic(err)
		}
		defer func(conn net.PacketConn) {
			if closeErr := conn.Close(); closeErr != nil {
				panic(closeErr)
			}
		}(c.conn)
	}

	// Set a handler for when a new remote track starts, this handler creates a gstreamer pipeline
	// for the given codec
	peerConnection.OnTrack(func(track *webrtc.Track, receiver *webrtc.RTPReceiver) {
		// Retrieve udp connection
		c, ok := udpConns[track.Kind().String()]
		if !ok {
			return
		}

		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		go func() {
			ticker := time.NewTicker(time.Second * 2)
			for range ticker.C {
				if rtcpErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: track.SSRC()}}); rtcpErr != nil {
					fmt.Println(rtcpErr)
				}
			}
		}()

		b := make([]byte, 1500)
		for {
			// Read
			n, readErr := track.Read(b)
			if readErr != nil {
				panic(readErr)
			}

			// Write
			if _, err = c.conn.Write(b[:n]); err != nil {
				// For this particular example, third party applications usually timeout after a short
				// amount of time during which the user doesn't have enough time to provide the answer
				// to the browser.
				// That's why, for this particular example, the user first needs to provide the answer
				// to the browser then open the third party application. Therefore we must not kill
				// the forward on "connection refused" errors
				if opError, ok := err.(*net.OpError); ok && opError.Err.Error() == "write: connection refused" {
					continue
				}
				panic(err)
			}
		}
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Logger.Info("Connection State has changed %s \n", connectionState.String())

		if connectionState == webrtc.ICEConnectionStateConnected {
			log.Logger.Info("Ctrl+C the remote client to stop the demo")
		} else if connectionState == webrtc.ICEConnectionStateFailed ||
			connectionState == webrtc.ICEConnectionStateDisconnected {
			log.Logger.Info("Done forwarding")
			cancel()
		}
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

	log.Logger.Debug("send sdp to send channel")
	s.SendCh <- utils.Encode(answer)

	if len(s.WebrtcSignal) > 0 {
		<-s.WebrtcSignal
	}

	<-ctx.Done()

	// Block forever
	select {
	case <- s.WebrtcSignal:
		log.Logger.Info("got exit webrtc signal, webrtc will exit")
		runtime.Goexit()
	}
}

func (s *Service) startGstream()  {
	var err		error
	var stdout 	bytes.Buffer
	var stderr 	bytes.Buffer

	// start gstreamer v4l2 video
	cmd := exec.Command( //nolint
		"gst-launch-1.0",
		"udpsrc", "port=5004",
		`caps="application/x-rtp,media=(string)video,clock-rate=(int)90000,encoding-name=(string)H264,payload=(int)96"`,
		"!", "rtph264depay",
		"!", "decodebin",
		"!", "videoconvert",
		"!", "autovideosink", "sync=false",
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
		panic(cmd.Stderr)
	}
}

func (s *Service) Run() {
	log.Logger.Info("controller service starting...")
	err := s.InitConn()
	if err != nil {
		log.Logger.Error(err)
	}

	go s.Send()
	go s.ServerRecv()

	go s.SignIn()
	go s.webrtc()
	go s.startGstream()

	select {}
}