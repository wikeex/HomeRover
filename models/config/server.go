package config

import (
	"github.com/vaughan0/go-ini"
	"strconv"
	"strings"
)

type ServerConfig struct {
	LocalPort		int			`json:"localPort"`
}

func GetDefaultServerConfig() ServerConfig {
	return ServerConfig{
		LocalPort:  	10006,
	}
}

func ServerConfigInit(filePath string) (serverConf ServerConfig, err error) {
	serverConf = GetDefaultServerConfig()

	conf, err := ini.Load(strings.NewReader(filePath))
	if err != nil {
		return ServerConfig{}, err
	}

	var (
		tempString		string
		ok				bool
		value			int
	)

	if tempString, ok = conf.Get("common", "localPort"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		serverConf.LocalPort = value
	}

	return
}
