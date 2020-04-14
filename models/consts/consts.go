package consts

// package type
const (
	ControllerServe = iota
	RoverServe
	ServerResp
	ControllerCmd
	RoverCmd
	ControllerVideo
	RoverVideo
	ControllerAudio
	RoverAudio
)

// transmission mode
const (
	HoldPunching = true
	ServerForwarding = false
)

// client state
const (
	Online = iota
	Offline
)