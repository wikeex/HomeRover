package main

import (
	"HomeRover/controller"
	"HomeRover/joystick"
	"HomeRover/log"
	"HomeRover/models/config"
)

func main() {
	log.Logger.Info("reading common config...")
	conf, err := config.CommonConfigInit("conf/controller.ini")
	if err != nil {
		log.Logger.Error(err)
	}

	log.Logger.Info("reading controller config")
	controllerConf, err := config.ControllerConfigInit("conf/controller.ini")
	if err != nil {
		log.Logger.Error(err)
	}

	joystickData := make(chan []byte, 1)
	js, err := joystick.NewJoystick(&controllerConf, joystickData)
	if err != nil {
		log.Logger.Error(err)
	}

	err = js.Init()
	if err != nil {
		log.Logger.Error(err)
	}

	log.Logger.Info("joystick task staring...")
	go js.Run()

	service, err := controller.NewService(&conf, &controllerConf, joystickData)
	if err != nil {
		log.Logger.Error(err)
	}

	service.Run()
}
