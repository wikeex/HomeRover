package config

import (
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
}

func GetDefaultControllerConfig() ControllerConfig {
	return ControllerConfig{
		ServerIP:		"140.143.99.31",
		ServerPort:		10006,
		LocalPort: 		18000,
		JoystickFreq: 	50,
		PackageLen: 	548,
	}
}

func ControllerConfigInit(filePath string) (controllerConfig ControllerConfig, err error) {
	controllerConf := GetDefaultControllerConfig()

	conf, err := ini.Load(strings.NewReader(filePath))
	if err != nil {
		return ControllerConfig{}, err
	}

	var (
		tempString		string
		ok				bool
		value			int
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

	return
}
