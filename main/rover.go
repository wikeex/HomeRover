package main

import (
	"HomeRover/models/config"
	"HomeRover/rover"
	"fmt"
)

func main()  {
	conf, err := config.CommonConfigInit("conf/rover.ini")
	if err != nil {
		fmt.Println(err)
	}

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
