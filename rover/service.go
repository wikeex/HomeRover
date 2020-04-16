package rover

import "HomeRover/base"

type Service struct {
	base.Service

	joystickData	chan []byte
}

