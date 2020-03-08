package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"udpTest"
)

func main()  {
	links := make(map[string]udpTest.ClientOnServer)

	serverAddrStr := "0.0.0.0" + ":" + strconv.Itoa(udpTest.SERVER_PORT)
	serverAddr, err := net.ResolveUDPAddr("udp", serverAddrStr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer conn.Close()

	receiveMap := udpTest.Data{}
	sendMap := make(map[string]interface{})
	sendMap["type"] = "serverResp"

	for {
		receiveData := make([]byte, 548)

		n, rAddr, err := conn.ReadFromUDP(receiveData)
		if err != nil && receiveData == nil {
			fmt.Println(err)
			continue
		}

		err = json.Unmarshal(receiveData[:n], &receiveMap)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(string(receiveData))

		// update Address and Destination when get heartbeat package
		if client, ok := links[receiveMap.ClientId]; ok{
			client.Address = rAddr.String()
			UpdateDest(&links)
		} else {
			links[receiveMap.ClientId] = udpTest.ClientOnServer{Id: receiveMap.ClientId, Address: rAddr.String()}
			UpdateDest(&links)
		}

		if len(links[receiveMap.ClientId].Destination) > 0 {
			sendMap["data"] = (links[receiveMap.ClientId].Destination)[0].Address
		}
		sendData, err := json.Marshal(sendMap)
		if err != nil {
			fmt.Println(err)
			continue
		}
		_, err = conn.WriteToUDP(sendData, rAddr)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("data: " + string(sendData) + " to: " + rAddr.String() + "recvData: " + string(receiveData))
	}
}

func UpdateDest(links *map[string]udpTest.ClientOnServer)  {
	for _, clientA := range *links {
		tempClient := udpTest.ClientOnServer{Id:clientA.Id, Address:clientA.Address}
		for _, clientB := range *links {
			if clientA.Id != clientB.Id {
				tempClient.Destination = append(tempClient.Destination, udpTest.Destination{Id:clientB.Id, Address:clientB.Address})
			}
		}
		(*links)[clientA.Id] = tempClient
	}
}