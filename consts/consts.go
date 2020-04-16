package consts

// client type
const (
	Rover = iota
	Controller
	Server
)

// channel
const (
	Cmd = iota
	Video
	Audio
	Service
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