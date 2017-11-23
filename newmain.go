// newmain.go
/*
	Gateway:
		rgb:
			<int>
		illumination:
			<int>


	Switch:
		status:
			"click"
			"double_click"
			"long_click_press"
			"long_click_release"

	Motion:
		status:
			"motion"
		no_motion:
			<sec>

	Sensor_ht:
		voltage:
			<int>
		temperature:
			<int>
		humidity:
			<int>

*/

package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	//	"strings"
	//	"sync"
	//	"time"
)

type Response struct {
	Cmd   string      `json:"cmd"`
	Model string      `json:"model"`
	Sid   string      `json:"sid"`
	Token string      `json:"token,omitempty"`
	IP    string      `json:"ip,omitempty"`
	Port  string      `json:"port,omitempty"`
	Data  interface{} `json:"data"`
}

type Request struct {
	Cmd string `json:"cmd"`
	Sid string `json:"sid,omitempty"`
}

type GatewayData struct {
	Ip           string `json:"ip"`
	Port         string `json:"port"`
	Rgb          int    `json:"rgb,omitempty"`
	Illumination int    `json:"illumination,omitempty"`
}

type MotionData struct {
	Voltage int    `json:"voltage"`
	Status  string `json:"status`
}

type MagnetData struct {
	Voltage int    `json:"voltage"`
	Status  string `json:"status`
}

type SwitchData struct {
	Voltage int    `json:"voltage"`
	Status  string `json:"status"`
}

type Sensor_htData struct {
	Voltage     int    `json:"voltage"`
	Temperature string `json:"temperature"`
	Humidity    string `json:"humidity"`
}

const (
	multicastIp     = "224.0.0.50"
	multicastPort   = "4321"
	maxDatagramSize = 1024
)

var (
	conn *net.UDPConn
)

func sendMessage(addr *net.UDPAddr, msg string, sid string) {

	var req []byte
	var err error

	if sid != "" {
		req, err = json.Marshal(Request{Cmd: msg, Sid: sid})
	} else {
		req, err = json.Marshal(Request{Cmd: msg})
	}

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Msg: %+v - Addr: %+v", string(req), addr)
	conn.WriteMsgUDP([]byte(req), nil, addr)
}

func serveMulticast(a string) {
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		log.Fatal(err)
	}
	conn, err = net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Panic(err)
	}
	go loopReadMulticast(conn, msgHandler)
	log.Printf("sending whois to %+v from %+v\n", addr, conn)
	sendMessage(addr, "whois", "")
}

func loopReadMulticast(conn *net.UDPConn, msgHandler func(resp *Response)) {
	conn.SetReadBuffer(maxDatagramSize)

	for {
		b := make([]byte, maxDatagramSize)
		n, _, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}

		log.Printf("read resp: %+v - size: %d", b, n)

		resp := Response{}
		err = json.Unmarshal(b[:n], &resp)
		if err != nil {
			log.Fatal(err)
		}
		msgHandler(&resp)
	}
}

func msgHandler(resp *Response) {
	switch resp.Cmd {
	case "iam":
		log.Printf("IAM: %+v", resp)
	case "heartbeat":
		log.Printf("HEARTBEAT: %+v", resp)
	case "get_id_list_ack":
		log.Printf("Get ACK: %+v\n", resp)
	case "read_ack":
		log.Printf("Read ACK: %+v", resp)
	default:
		log.Printf("DEFAULT: %+v", resp)
	}
}

func main() {
	f, err := os.OpenFile("xiaomi.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
	log.Println("Starting handler...")

	serveMulticast(multicastIp + ":" + multicastPort)

	for {
	}
}
