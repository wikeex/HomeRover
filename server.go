package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
	"udpTest/config"
)

func main()  {
	address := "0.0.0.0" + ":" + strconv.Itoa(config.SERVER_PORT)
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer conn.Close()

	receiveData := make([]byte, config.SERVER_RECV_LEN)
	sendMap := make(map[string]string)
	sendMap["type"] = "addr"

	for {
		_, rAddr, err := conn.ReadFromUDP(receiveData)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if rAddr != nil && sendMap["data"] != rAddr.String() {
			sendMap["data"] = rAddr.String()

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

			fmt.Println("Resent: " + string(sendData) + " to: " + rAddr.String() + "recvData: " + string(receiveData))
		}
		time.Sleep(time.Second)
	}
}
