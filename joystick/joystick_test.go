package joystick

import (
	"HomeRover/models/config"
	"fmt"
	"testing"
)

func TestReadJoystick(t *testing.T) {

	controllerConf, err := config.ControllerConfigInit("../conf/controller.ini")
	if err != nil {
		fmt.Println(err)
	}

	joystickData := make(chan []byte, 1)
	joystick, err := NewJoystick(&controllerConf, joystickData)
	if err != nil {
		t.Error(err)
	}

	err = joystick.Init()
	if err != nil {
		t.Fail()
	}

	data, err := joystick.ReadOnce()
	if err != nil {
		t.Fail()
	}

	if data != nil {
		t.Log("PASS")
	}

	go joystick.Run()

	for i := 0; i < 10; i++ {
		t.Log(<- joystickData)
	}
	t.Log("PASS")
}
