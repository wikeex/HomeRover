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
	"udpTest/config"
)

var cond *sync.Cond
var ClientId string

func init()  {
	cond = sync.NewCond(&sync.Mutex{})
	rand.Seed(time.Now().UnixNano())
	ClientId = strconv.Itoa(rand.Int())
}

func send(conn *net.UDPConn, addr **net.UDPAddr, dataCh chan []byte)  {
	cond.L.Lock()
	cond.Wait()
	defer cond.L.Unlock()
	for {
		data := <- dataCh
		_, err := conn.WriteToUDP(data, *addr)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("send data: " + string(data))
		time.Sleep(time.Second)
	}
}

func server(conn *net.UDPConn, beatCount *int) {
	sender := udpTest.Data{Type:"serverReq", ClientId:ClientId}
	sendData, err := json.Marshal(sender)
	if err != nil {
		fmt.Println(err)
	}

	addrStr := config.SERVER_IP + ":" + strconv.Itoa(config.SERVER_PORT)
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

func receive(conn *net.UDPConn, beatCount *int, sendDataCh chan []byte, addr **net.UDPAddr, isReady *bool)  {

	receiver := udpTest.Data{}
	sender := udpTest.Data{ClientId:ClientId}

	startTime := time.Now().UnixNano()

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
			*beatCount++

			*addr, err = net.ResolveUDPAddr("udp", receiver.Data)
			if err != nil {
				fmt.Println(err)
			}
			cond.Broadcast()
		case "probe":
			sender.Type = "probeResp"
			sendData, _ := json.Marshal(sender)
			_, err = conn.WriteToUDP(sendData, *addr)
		case "probeResp":
			*isReady = true
		case "send":
			sender.Type = "receipt"
			sender.Data = receiver.Data
			sendData, _ := json.Marshal(sender)
			_, err = conn.WriteToUDP(sendData, *addr)
		case "receipt":
			if receiver.Data == sender.Data {
				fmt.Print("耗时：")
				fmt.Println(time.Now().UnixNano() - startTime)
			}
			sender.Type = "send"
			sender.Data = strconv.Itoa(rand.Int())
			sendData, _ := json.Marshal(sender)
			sendDataCh <- sendData
			startTime = time.Now().UnixNano()
		}

	}

}

func sendProbe(conn *net.UDPConn, addr **net.UDPAddr, isReady *bool)  {
	cond.L.Lock()
	cond.Wait()
	defer cond.L.Unlock()

	sender := udpTest.Data{Type:"probe", ClientId:ClientId}
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

func inputData(dataCh chan []byte)  {
	sender := udpTest.Data{Type:"send", ClientId:ClientId}
	var keyBoardData string
	_, err := fmt.Scanln(&keyBoardData)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("get keyboard input: " + keyBoardData)

	sender.Data = strconv.Itoa(rand.Int())
	sendData, _ := json.Marshal(sender)
	dataCh <- sendData
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
	sendDataCh := make(chan []byte)
	isReady := false

	go server(conn, &beatCount)
	go send(conn, &addr, sendDataCh)
	go sendProbe(conn, &addr, &isReady)
	go inputData(sendDataCh)
	receive(conn, &beatCount, sendDataCh, &addr, &isReady)
}