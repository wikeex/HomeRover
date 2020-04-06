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

func send(conn *net.UDPConn, dataCh chan string)  {

	sender := models.Data{ClientId: ClientId}
	addr, err := net.ResolveUDPAddr("udp", HomeRover.SERVER_IP + ":" + strconv.Itoa(HomeRover.SERVER_PORT))
	if err != nil {
		fmt.Println(err)
	}

	sender.Type = "send"

	for {
		sender.Timestamp = time.Now().UnixNano()
		sender.Data = <- dataCh
		data, _ := json.Marshal(sender)
		_, err := conn.WriteToUDP(data, addr)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("send data: " + string(data))
		time.Sleep(time.Second)
	}
}

func server(conn *net.UDPConn) {
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

		time.Sleep(time.Second)
	}
}

func receive(conn *net.UDPConn, dataMap *map[string]interface{}, isReady *bool)  {

	receiver := models.Data{}
	sender := models.Data{ClientId: ClientId}
	addr, err := net.ResolveUDPAddr("udp", HomeRover.SERVER_IP + ":" + strconv.Itoa(HomeRover.SERVER_PORT))
	if err != nil {
		fmt.Println(err)
	}

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
		case "send":
			sender.Type = "receipt"
			sender.Data = receiver.Data
			sender.Timestamp = receiver.Timestamp
			sendData, _ := json.Marshal(sender)
			_, err = conn.WriteToUDP(sendData, addr)
		case "receipt":
			// 计算丢包率
			mux.Lock()
			if _, ok := (*dataMap)[receiver.Data]; !ok {
				break
			}
			delete(*dataMap, receiver.Data)
			missingRatio := float32(len(*dataMap) - 11) / float32(((*dataMap)["count"]).(int64))
			mux.Unlock()
			fmt.Print("耗时：")
			fmt.Println(time.Now().UnixNano() - receiver.Timestamp)
			fmt.Printf("丢包率：%.6f%%", missingRatio * 100)
			fmt.Println("")
		case "conn":
			sender.Type = "connResp"
			sendData, _ := json.Marshal(sender)
			_, err = conn.WriteToUDP(sendData, addr)
		case "connResp":
			mux.Lock()
			*isReady = true
			mux.Unlock()
		}

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

func establishConn(conn *net.UDPConn, isReady *bool)  {
	sender := models.Data{ClientId: ClientId}
	addr, err := net.ResolveUDPAddr("udp", HomeRover.SERVER_IP + ":" + strconv.Itoa(HomeRover.SERVER_PORT))
	if err != nil {
		fmt.Println(err)
	}

	sender.Type = "conn"

	for !(*isReady) {
		sender.Timestamp = time.Now().UnixNano()
		data, _ := json.Marshal(sender)
		_, err := conn.WriteToUDP(data, addr)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("establish connection data: " + string(data))
		time.Sleep(time.Second)
	}
}

func main()  {
	localAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:18800")

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer conn.Close()

	sendDataCh := make(chan string, 10)
	dataMap := make(map[string]interface{})
	isReady := false

	go server(conn)
	go receive(conn, &dataMap, &isReady)
	establishConn(conn, &isReady)
	go send(conn, sendDataCh)

	go inputData()
	genData(sendDataCh, &dataMap)

}