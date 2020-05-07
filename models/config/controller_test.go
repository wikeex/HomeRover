package config

import "testing"

func TestControllerConfigInit(t *testing.T) {
	conf, err := ControllerConfigInit("../../conf/controller.ini")
	if err != nil {
		t.Error(err)
	}

	if conf.JoystickFreq == 50 {
		t.Log("PASS")
	}
}
