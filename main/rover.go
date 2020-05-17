package main

import (
	"HomeRover/log"
	"HomeRover/models/config"
	"HomeRover/rover"
	"fmt"
)

func main()  {
	log.Logger.Info("reading common config...")
	conf, err := config.CommonConfigInit("conf/rover.ini")
	if err != nil {
		fmt.Println(err)
	}

	log.Logger.Info("reading rover config...")
	roverConf, err := config.RoverConfigInit("conf/rover.ini")
	if err != nil {
		fmt.Println(err)
	}

	service, err := rover.NewService(&conf, &roverConf)
	if err != nil {
		fmt.Println(err)
	}

	service.Run()
}
