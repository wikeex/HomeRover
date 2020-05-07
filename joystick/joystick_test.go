package test

import (
	"HomeRover/joystick"
	"testing"
)

func TestReadJoystick(t *testing.T) {
	controller, err := joystick.GetJoystick()
	if err != nil {
		t.Fail()
	}

	data, err := joystick.ReadOnce(controller)
	if err != nil {
		t.Fail()
	}

	if data != nil {
		t.Log("PASS")
	}
}
