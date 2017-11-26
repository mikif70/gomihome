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
	//	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	//	"strings"
	"sync"
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

var (
	udp     bool
	conn    *net.UDPConn
	gateway *Gateway
	wg      sync.WaitGroup
)

func main() {
	f, err := os.OpenFile("xiaomi.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	log.Println("Starting handler...")

	gateway = &Gateway{}

	multicast = true
	wg.Add(1)
	serveMulticast(multicastIp + ":" + multicastPort)
	wg.Wait()
	log.Printf("Gateway: %+v", gateway)

	udp = true
	wg.Add(1)
	gateway.serveUDP()
	gateway.discoverDevs()
	wg.Wait()
}
