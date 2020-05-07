package config

import (
	"github.com/vaughan0/go-ini"
	"strconv"
)

type ServerConfig struct {
	ServicePort		int		`json:"servicePort"`
	ForwardPort		int		`json:"forwardPort"`
	PackageLen		int		`json:"packageLen"`
}

func GetDefaultServerConfig() ServerConfig {
	return ServerConfig{
		ServicePort:  	10006,
		ForwardPort: 	10007,
		PackageLen:  	548,
	}
}

func ServerConfigInit(filePath string) (serverConf ServerConfig, err error) {
	serverConf = GetDefaultServerConfig()

	conf, err := ini.LoadFile(filePath)
	if err != nil {
		return ServerConfig{}, err
	}

	var (
		tempString		string
		ok				bool
		value			int
	)

	if tempString, ok = conf.Get("common", "servicePort"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		serverConf.ServicePort = value
	}

	if tempString, ok = conf.Get("common", "packageLen"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		serverConf.PackageLen = value
	}

	if tempString, ok = conf.Get("common", "forwardPort"); ok {
		value, err = strconv.Atoi(tempString)
		if err != nil {
			return
		}
		serverConf.ForwardPort = value
	}

	return
}
