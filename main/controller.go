package main

import (
	"HomeRover/joystick"
	"HomeRover/models/config"
	"fmt"
)

func main() {
	conf, err := config.ControllerConfigInit("./controller.ini")
	if err != nil {
		fmt.Println(err)
	}

	joystickData := make(chan []byte, 1)
	js, err := joystick.NewJoystick(&conf, &joystickData)
	if err != nil {
		fmt.Println(err)
	}

	err = js.Init()
	if err != nil {
		fmt.Println(err)
	}

	go js.Run()

}
