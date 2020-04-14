package group

import (
	"HomeRover/models/mode"
	"HomeRover/models/server"
)

type Group struct {
	Id 			uint16

	Rover 		server.Client
	Controller 	server.Client

	Trans 		*mode.Trans
}


