package udpTest

type Data struct {
	Type string `json:"type"`
	ClientId string `json:"ClientId"`
	GroupId string `json:"groupId"`
	PackageId string `json:"packageId"`
	Data string `json:"data"`
}

type Destination struct {
	Id string `json:"id"`
	Address string `json:"address"`
}


type ClientOnServer struct {
	Id string `json:"Id"`
	Destination []Destination `json:"destination"`
	Address string `json:"address"`
}


