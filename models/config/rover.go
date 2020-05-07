package config

import (
	"github.com/vaughan0/go-ini"
	"strconv"
)

type RoverConfig struct {
	CmdServicePort		int		`json:"cmdServicePort"`
}

func GetDefaultRoverConfig() RoverConfig {
	return RoverConfig{
		CmdServicePort: 10008,
	}
}

func RoverConfigInit(filePath string) (roverConfig RoverConfig, err error) {
	roverConfig = GetDefaultRoverConfig()

	conf, err := ini.LoadFile(filePath)
	if err != nil {
		return RoverConfig{}, err
	}

	var (
		tempString		string
		ok				bool
		value			int
	)

	if tempString, ok = conf.Get("common", "cmdServicePort"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		roverConfig.CmdServicePort = value
	}

	return
}