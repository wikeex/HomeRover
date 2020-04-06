package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
	"udpTest"
	"udpTest/models"
)

var ClientId string
var mux = sync.Mutex{}

func init()  {
	rand.Seed(time.Now().UnixNano())
	ClientId = strconv.Itoa(rand.Int())
}

func send(conn *net.UDPConn, addr **net.UDPAddr, dataCh chan string)  {

	sender := models.Data{ClientId: ClientId}

	sender.Type = "send"

	for {
		sender.Timestamp = time.Now().UnixNano()
		sender.Data = <- dataCh
		data, _ := json.Marshal(sender)
		_, err := conn.WriteToUDP(data, *addr)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("send data: " + string(data))
		time.Sleep(time.Second)
	}
}

func server(conn *net.UDPConn, beatCount *int) {
	sender := models.Data{Type: "serverReq", ClientId:ClientId}
	sendData, err := json.Marshal(sender)
	if err != nil {
		fmt.Println(err)
	}

	addrStr := HomeRover.SERVER_IP + ":" + strconv.Itoa(HomeRover.SERVER_PORT)
	addr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		fmt.Println(err)
	}

	for {
		_, err = conn.WriteToUDP(sendData, addr)
		if err != nil {
			fmt.Println(err)
		}

		*beatCount--
		time.Sleep(time.Second)
	}
}

func receive(conn *net.UDPConn, addr **net.UDPAddr, isReady *bool, dataMap *map[string]interface{})  {

	receiver := models.Data{}
	sender := models.Data{ClientId: ClientId}

	for {
		receiveData := make([]byte, 548)

		n, _, err := conn.ReadFromUDP(receiveData)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = json.Unmarshal(receiveData[:n], &receiver)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("receive data: " + string(receiveData))

		switch receiver.Type {
		case "serverResp":
			if receiver.Data == "" {
				break
			}

			*addr, err = net.ResolveUDPAddr("udp", receiver.Data)
			if err != nil {
				fmt.Println(err)
			}
		case "probe":
			sender.Type = "probeResp"
			sendData, _ := json.Marshal(sender)
			_, err = conn.WriteToUDP(sendData, *addr)
		case "probeResp":
			*isReady = true
		case "send":
			sender.Type = "receipt"
			sender.Data = receiver.Data
			sender.Timestamp = receiver.Timestamp
			sendData, _ := json.Marshal(sender)
			_, err = conn.WriteToUDP(sendData, *addr)
		case "receipt":
			// 计算丢包率
			mux.Lock()
			delete(*dataMap, receiver.Data)
			missingRatio := float32(len(*dataMap) - 11) / float32(((*dataMap)["count"]).(int64))
			mux.Unlock()
			fmt.Print("耗时：")
			fmt.Println(time.Now().UnixNano() - receiver.Timestamp)
			fmt.Printf("丢包率：%.6f%%", missingRatio * 100)
			fmt.Println("")
		}

	}

}

func sendProbe(conn *net.UDPConn, addr **net.UDPAddr, isReady *bool)  {
	sender := models.Data{Type: "probe", ClientId:ClientId}
	sendData, err := json.Marshal(sender)
	if err != nil {
		fmt.Println(err)
	}

	for !*isReady {
		_, err = conn.WriteToUDP(sendData, *addr)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("sendProbe: " + string(sendData))

		time.Sleep(50 * time.Millisecond)
	}
}

func genData(sendDataCh chan string, dataMap *map[string]interface{})  {
	var count int64
	for {
		data := strconv.Itoa(rand.Int())
		sendDataCh <- data
		mux.Lock()
		(*dataMap)[data] = true
		(*dataMap)["count"] = count
		mux.Unlock()
		count++
	}
}

func inputData()  {
	var keyBoardData string

	_, err := fmt.Scanln(&keyBoardData)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("get keyboard input: " + keyBoardData)

}


func main()  {
	localAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:18801")

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer conn.Close()

	var addr *net.UDPAddr

	beatCount := 10
	sendDataCh := make(chan string, 10)
	isReady := false
	dataMap := make(map[string]interface{})

	go server(conn, &beatCount)
	go receive(conn, &addr, &isReady, &dataMap)
	sendProbe(conn, &addr, &isReady)
	go send(conn, &addr, sendDataCh)

	go inputData()
	genData(sendDataCh, &dataMap)

}