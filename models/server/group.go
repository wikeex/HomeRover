package server

import (
	"HomeRover/models/client"
)

type Group struct {
	Id 			uint16

	Rover      client.Client
	Controller client.Client
}


