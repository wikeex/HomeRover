package server

import (
	"HomeRover/models/client"
	"HomeRover/models/mode"
)

type Group struct {
	Id 			uint16

	Rover      client.Client
	Controller client.Client

	Trans 		*mode.Trans
}


