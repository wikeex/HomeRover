package config

import (
	"HomeRover/models/mode"
	"github.com/vaughan0/go-ini"
	"strconv"
	"strings"
)

type ControllerConfig struct {
	ServerIP		string		`json:"serverIp"`
	ServerPort		int			`json:"serverPort"`
	LocalPort		int			`json:"localPort"`
	JoystickFreq	int			`json:"joystickFreq"`
	PackageLen		int			`json:"packageLen"`
	ControllerId	int			`json:"controllerId"`
	GroupId			int			`json:"groupId"`

	Trans 			mode.Trans
	CmdMode   		mode.TransMode 	`json:"cmd"`
	VideMode  		mode.TransMode 	`json:"video"`
	AudioMode 		mode.TransMode 	`json:"audio"`
}

func GetDefaultControllerConfig() ControllerConfig {
	return ControllerConfig{
		ServerIP:		"140.143.99.31",
		ServerPort:		10006,
		LocalPort: 		18000,
		JoystickFreq: 	50,
		PackageLen: 	548,

		CmdMode: 		false,
		VideMode: 		false,
		AudioMode: 		false,
	}
}

func ControllerConfigInit(filePath string) (controllerConf ControllerConfig, err error) {
	controllerConf = GetDefaultControllerConfig()

	conf, err := ini.Load(strings.NewReader(filePath))
	if err != nil {
		return ControllerConfig{}, err
	}

	var (
		tempString		string
		ok				bool
		value			int
		boolValue		bool
	)

	if tempString, ok = conf.Get("common", "serverIp"); ok {
		controllerConf.ServerIP = tempString
	}

	if tempString, ok = conf.Get("common", "serverPort"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		controllerConf.ServerPort = value
	}

	if tempString, ok = conf.Get("common", "localPort"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		controllerConf.LocalPort = value
	}

	if tempString, ok = conf.Get("common", "joystickFreq"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		controllerConf.JoystickFreq = value
	}

	if tempString, ok = conf.Get("common", "packageLen"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		controllerConf.PackageLen = value
	}

	if tempString, ok = conf.Get("common", "ControllerId"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		controllerConf.ControllerId = value
	} else {
		panic("'ControllerId' variable missing from 'common' section")
	}

	if tempString, ok = conf.Get("common", "groupId"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		controllerConf.GroupId = value
	} else {
		panic("'groupId' variable missing from 'common' section")
	}

	if tempString, ok = conf.Get("mode", "cmd"); ok {
		boolValue, err = strconv.ParseBool(tempString)
		if err != nil {
			return
		}
		controllerConf.CmdMode = mode.TransMode(boolValue)
	}

	if tempString, ok = conf.Get("mode", "video"); ok {
		boolValue, err = strconv.ParseBool(tempString)
		if err != nil {
			return
		}
		controllerConf.VideMode = mode.TransMode(boolValue)
	}

	if tempString, ok = conf.Get("mode", "audio"); ok {
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
