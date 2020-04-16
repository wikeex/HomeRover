package config

import (
	"github.com/vaughan0/go-ini"
	"strconv"
	"strings"
)

type RoverConfig struct {
	ServerIP		string		`json:"serverIp"`
	ServerPort		int			`json:"serverPort"`
	LocalPort		int			`json:"localPort"`
	PackageLen		int			`json:"packageLen"`
	Id				int			`json:"id"`
	GroupId			int			`json:"groupId"`
}

func GetDefaultRoverConfig() RoverConfig {
	return RoverConfig{
		ServerIP:   "140.143.99.31",
		ServerPort: 10006,
		LocalPort:  18000,
		PackageLen: 548,
	}
}

func RoverConfigInit(filePath string) (roverConfig RoverConfig, err error) {
	roverConfig = GetDefaultRoverConfig()

	conf, err := ini.Load(strings.NewReader(filePath))
	if err != nil {
		return RoverConfig{}, err
	}

	var (
		tempString		string
		ok				bool
		value			int
	)

	if tempString, ok = conf.Get("common", "serverIp"); ok {
		roverConfig.ServerIP = tempString
	}

	if tempString, ok = conf.Get("common", "serverPort"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		roverConfig.ServerPort = value
	}

	if tempString, ok = conf.Get("common", "localPort"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		roverConfig.LocalPort = value
	}

	if tempString, ok = conf.Get("common", "packageLen"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		roverConfig.PackageLen = value
	}

	if tempString, ok = conf.Get("common", "id"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		roverConfig.Id = value
	} else {
		panic("'ControllerId' variable missing from 'common' section")
	}

	if tempString, ok = conf.Get("common", "groupId"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		roverConfig.GroupId = value
	} else {
		panic("'groupId' variable missing from 'common' section")
	}

	return
}