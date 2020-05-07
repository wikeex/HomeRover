package config

import (
	"HomeRover/models/mode"
	"github.com/vaughan0/go-ini"
	"strconv"
)

type ControllerConfig struct {
	JoystickFreq	int				`json:"joystickFreq"`
	Trans 			mode.Trans
	CmdMode   		mode.TransMode 	`json:"cmd"`
	VideMode  		mode.TransMode 	`json:"video"`
	AudioMode 		mode.TransMode 	`json:"audio"`
}

func GetDefaultControllerConfig() ControllerConfig {
	return ControllerConfig{
		JoystickFreq: 	50,

		CmdMode: 		false,
		VideMode: 		false,
		AudioMode: 		false,
	}
}

func ControllerConfigInit(filePath string) (controllerConf ControllerConfig, err error) {
	controllerConf = GetDefaultControllerConfig()

	conf, err := ini.LoadFile(filePath)
	if err != nil {
		return ControllerConfig{}, err
	}

	var (
		tempString		string
		ok				bool
		value			int
		boolValue		bool
	)

	if tempString, ok = conf.Get("controller", "joystickFreq"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		controllerConf.JoystickFreq = value
	}

	if tempString, ok = conf.Get("controller", "cmd"); ok {
		boolValue, err = strconv.ParseBool(tempString)
		if err != nil {
			return
		}
		controllerConf.CmdMode = mode.TransMode(boolValue)
	}

	if tempString, ok = conf.Get("controller", "video"); ok {
		boolValue, err = strconv.ParseBool(tempString)
		if err != nil {
			return
		}
		controllerConf.VideMode = mode.TransMode(boolValue)
	}

	if tempString, ok = conf.Get("controller", "audio"); ok {
		boolValue, err = strconv.ParseBool(tempString)
		if err != nil {
			return
		}
		controllerConf.AudioMode = mode.TransMode(boolValue)
	}

	controllerConf.Trans = mode.Trans{
		Cmd: 		controllerConf.CmdMode,
		Video:		controllerConf.VideMode,
		Audio: 		controllerConf.AudioMode,
	}

	return
}
