package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type (
	Response struct {
		Cmd   string      `json:"cmd"`
		Model string      `json:"model"`
		Sid   string      `json:"sid"`
		Token string      `json:"token,omitempty"`
		IP    string      `json:"ip,omitempty"`
		Port  string      `json:"port,omitempty"`
		Data  interface{} `json:"data"`
	}

	Request struct {
		Cmd string `json:"cmd"`
		Sid string `json:"sid,omitempty"`
	}

	Gateway struct {
		Addr  *net.UDPAddr
		Sid   string
		Token string
	}

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

	Device struct {
		Model        string
		Sid          string
		Name         string
		Voltage      int
		Status       string
		Temperature  string
		Humidity     string
		Illumination int
		Nomotion     int
		Rgb          int
	}

	MotionData struct {
		Voltage int    `json:"voltage"`
		Status  string `json:"status`
	}

	MagnetData struct {
		Voltage int    `json:"voltage"`
		Status  string `json:"status`
	}

	SwitchData struct {
		Voltage int    `json:"voltage"`
		Status  string `json:"status"`
	}

	Sensor_htData struct {
		Voltage     int    `json:"voltage"`
		Temperature string `json:"temperature"`
		Humidity    string `json:"humidity"`
	}

	object map[string]interface{}
)

var (
	conn     *net.UDPConn
	gateways *Gateway
	wg       sync.WaitGroup
	whois    = 0
	devices  = make(map[string]*Device)
)

const (
	multicastIp     = "224.0.0.50"
	gatewayIp       = "192.168.1.150"
	multicastPort   = "9898"
	maxDatagramSize = 8192
)

func (this *Gateway) sendMessage(msg string, sid string) {
	sendMessage(this.Addr, msg, sid)
}

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
	log.Println(string(req))
	log.Println(addr)
	conn.WriteMsgUDP([]byte(req), nil, addr)
}

func connHandler() {
	pingAddr, err := net.ResolveUDPAddr("udp", multicastIp+":4321")
	if err != nil {
		log.Fatal(err)
	}
	sendMessage(pingAddr, "whois", "")
}

func msgHandler(resp *Response) {
	switch resp.Cmd {
	case "iam":
		whois++
		log.Printf("IAM: %+v", resp)
		gateways = &Gateway{
			Sid: resp.Sid,
		}
		gwaddr, err := net.ResolveUDPAddr("udp", resp.IP+":"+multicastPort)
		if err != nil {
			log.Fatal(err)
		}
		gateways.Addr = gwaddr
		gateways.Sid = resp.Sid
		gateways.Token = resp.Token

		if whois <= 1 {
			log.Printf("getting ID\n")
			gateways.sendMessage("get_id_list", "")
		}

	case "heartbeat":
		log.Printf("HEARTBEAT: %+v", resp)
	case "get_id_list_ack":
		log.Printf("Get ACK: %+v\n", resp)
		log.Printf("Data: %+v\n", resp.Data)
		retval := strings.Split(resp.Data.(string), ",")
		r := strings.NewReplacer("\"", "", "[", "", "]", "")
		for i := range retval {
			ns := r.Replace(retval[i])
			log.Printf("Data: %d - %s\n", i, ns)
			gateways.sendMessage("read", ns)
		}
	case "read_ack":
		log.Printf("Read ACK: %+v", resp)
		switch resp.Model {
		case "motion":
			data := MotionData{}
			err := json.Unmarshal([]byte(resp.Data.(string)), &data)
			if err != nil {
				log.Fatal(err)
			}
			updateDevice(resp.Sid, resp.Model, data)
		case "sensor_ht":
			data := Sensor_htData{}
			err := json.Unmarshal([]byte(resp.Data.(string)), &data)
			if err != nil {
				log.Fatal(err)
			}
			updateDevice(resp.Sid, resp.Model, data)
		case "switch":
			data := SwitchData{}
			err := json.Unmarshal([]byte(resp.Data.(string)), &data)
			if err != nil {
				log.Fatal(err)
			}
			updateDevice(resp.Sid, resp.Model, data)
		case "magnet":
			data := MagnetData{}
			err := json.Unmarshal([]byte(resp.Data.(string)), &data)
			if err != nil {
				log.Fatal(err)
			}
			updateDevice(resp.Sid, resp.Model, data)
		}

	default:
		log.Printf("DEFAULT: %+v", resp)
	}
}

func updateDevice(sid string, model string, data interface{}) {
	if _, ok := devices[sid]; !ok {
		devices[sid] = &Device{
			Sid:   sid,
			Model: model,
		}
	}
	switch model {
	case "sensor_ht":
		log.Printf("Before Update: %+v", devices[sid])
		devices[sid].Voltage = data.(Sensor_htData).Voltage
		devices[sid].Temperature = data.(Sensor_htData).Temperature
		devices[sid].Humidity = data.(Sensor_htData).Humidity
		log.Printf("After Update: %+v", devices[sid])
	}
}

func serveMulticastUDP(a string) {
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		log.Fatal(err)
	}
	//	conn, err = net.ListenMulticastUDP("udp", nil, addr)
	conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Panic(err)
	}
}

func loopReadUdp(conn *net.UDPConn, msgHandler func(resp *Response)) {
	conn.SetReadBuffer(maxDatagramSize)
	for {
		b := make([]byte, maxDatagramSize)
		n, _, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}

		resp := Response{}
		err = json.Unmarshal(b[:n], &resp)
		if err != nil {
			log.Fatal(err)
		}
		msgHandler(&resp)
	}
}

func main() {

	var err error

	f, err := os.OpenFile("xiaomi.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	mw := io.MultiWriter(os.Stdout, f)

	log.SetOutput(mw)

	gateways = &Gateway{}

	log.Println("Starting handler...")
	serveMulticastUDP(multicastIp + ":" + multicastPort)

	go loopReadUdp(conn, msgHandler)

	gateways.Addr, err = net.ResolveUDPAddr("udp", multicastIp+":4321") //+multicastPort)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("sending whois to %+v from %+v\n", gateways, conn)
	gateways.sendMessage("whois", "")

	for {
		if len(devices) != 0 {
			for k, v := range devices {
				switch v.Model {
				case "sensor_ht":
					log.Printf("k: %s - v: %+v", k, v)
				case "motion":
				case "magnet":
				case "switch":
				}
			}
		}
		time.Sleep(time.Minute * 1)
	}
}
