package HomeRover

import "net"

type Data struct {
	Type string `json:"type"`
	ClientId string `json:"ClientId"`
	GroupId string `json:"groupId"`
	PackageId string `json:"packageId"`
	Data string `json:"data"`
	Timestamp int64 `json:"timestamp"`
}

type Destination struct {
	Id string `json:"id"`
	Address *net.UDPAddr `json:"address"`
}


type ClientOnServer struct {
	Id string `json:"Id"`
	Destination []Destination `json:"destination"`
	Address string `json:"address"`
}


