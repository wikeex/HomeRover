package main

import (
	"encoding/json"
"fmt"
"net"
"os"
"strconv"
	"udpTest"
	"udpTest/models"
)

func main()  {
	links := make(map[string]models.ClientOnServer)

	serverAddrStr := "0.0.0.0" + ":" + strconv.Itoa(HomeRover.SERVER_PORT)
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

	receiveMap := models.Data{}
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

		switch receiveMap.Type {
		case "serverReq":
			// update Address and Destination when get heartbeat package
			if client, ok := links[receiveMap.ClientId]; ok{
				client.Address = rAddr.String()
				UpdateDest(&links)
			} else {
				links[receiveMap.ClientId] = models.ClientOnServer{Id: receiveMap.ClientId, Address: rAddr.String()}
				UpdateDest(&links)
			}

			if len(links[receiveMap.ClientId].Destination) > 0 {
				sendMap["client"] = (links[receiveMap.ClientId].Destination)[0].Address.String()
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
		default:
			if len(links[receiveMap.ClientId].Destination) > 0 {
				_, err = conn.WriteToUDP(receiveData[:n], (links[receiveMap.ClientId].Destination)[0].Address)
				fmt.Println(links[receiveMap.ClientId].Destination)
				fmt.Println("forwarding...")
			}
		}
	}
}

func UpdateDest(links *map[string]models.ClientOnServer)  {
	for _, clientA := range *links {
		tempClient := models.ClientOnServer{Id: clientA.Id, Address:clientA.Address}
		for _, clientB := range *links {
			if clientA.Id != clientB.Id {
				addr, err := net.ResolveUDPAddr("udp", clientB.Address)
				if err != nil {
					fmt.Println(err)
				}
				tempClient.Destination = append(tempClient.Destination, models.Destination{Id: clientB.Id, Address:addr})
			}
		}
		(*links)[clientA.Id] = tempClient
	}
}