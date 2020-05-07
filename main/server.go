package main

import (
	"HomeRover/models/config"
	"HomeRover/server"
	"fmt"
)

func main()  {
	conf, err := config.ServerConfigInit("./server.ini")
	if err != nil {
		fmt.Println(err)
	}

	service, err := server.NewService(&conf)
	if err != nil {
		fmt.Println(err)
	}

	service.Run()
}
