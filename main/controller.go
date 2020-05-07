package main

import (
	"HomeRover/controller"
	"HomeRover/joystick"
	"HomeRover/models/config"
	"fmt"
)

func main() {
	conf, err := config.CommonConfigInit("conf/controller.ini")
	if err != nil {
		fmt.Println(err)
	}

	controllerConf, err := config.ControllerConfigInit("conf/controller.ini")
	if err != nil {
		fmt.Println(err)
	}

	joystickData := make(chan []byte, 1)
	js, err := joystick.NewJoystick(&controllerConf, joystickData)
	if err != nil {
		fmt.Println(err)
	}

	err = js.Init()
	if err != nil {
		fmt.Println(err)
	}

	go js.Run()

	service, err := controller.NewService(&conf, &controllerConf, joystickData)
	if err != nil {
		fmt.Println(err)
	}

	service.Run()
}
