package config

import (
	"github.com/vaughan0/go-ini"
	"strings"
)

type ServerConfig struct {
	ServicePort		string		`json:"servicePort"`
	ForwardPort		string		`json:"forwardPort"`
}

func GetDefaultServerConfig() ServerConfig {
	return ServerConfig{
		ServicePort:  	"10006",
		ForwardPort: 	"10007",
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
	)

	if tempString, ok = conf.Get("common", "localPort"); ok {
		serverConf.ServicePort = tempString
	}

	if tempString, ok = conf.Get("common", "forwardPort"); ok {
		serverConf.ForwardPort = tempString
	}

	return
}
