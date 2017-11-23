package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"reflect"
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
		Ip    string
		Port  string
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

	GatewayData struct {
		Ip           string `json:"ip"`
		Port         string `json:"port"`
		Rgb          int    `json:"rgb,omitempty"`
		Illumination int    `json:"illumination,omitempty"`
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
	mconn    *net.UDPConn
	gateways *Gateway
	wg       sync.WaitGroup
	whois    = 0
	devices  = make(map[string]object)
)

const (
	multicastIp     = "224.0.0.50"
	multicastPort   = "4321"
	gatewayIp       = "192.168.1.150"
	gatewayPort     = "9898"
	maxDatagramSize = 1024
)

func (gw *Gateway) sendMessage(msg string, sid string) {
	sendMessage(gw.Addr, msg, sid)
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

/*
func connHandler() {
	pingAddr, err := net.ResolveUDPAddr("udp", multicastIp+":4321")
	if err != nil {
		log.Fatal(err)
	}
	sendMessage(pingAddr, "whois", "")
}
*/

func msgHandler(resp *Response) {
	switch resp.Cmd {
	case "iam":
		whois++
		log.Printf("IAM: %+v", resp)
		gateways = &Gateway{
			Sid: resp.Sid,
		}
		gwaddr, err := net.ResolveUDPAddr("udp", resp.IP+":"+resp.Port)
		if err != nil {
			log.Fatal(err)
		}
		gateways.Addr = gwaddr
		gateways.Sid = resp.Sid
		gateways.Token = resp.Token
		gateways.Ip = resp.IP
		gateways.Port = resp.Port

		if whois <= 1 {
			log.Printf("getting ID\n")
			gateways.sendMessage("get_id_list", "")
		}

	case "heartbeat":
		log.Printf("HEARTBEAT: %+v", resp)
		updateDevice(resp.Sid, resp.Model, resp.Data)
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
		updateDevice(resp.Sid, resp.Model, resp.Data)

	default:
		log.Printf("DEFAULT: %+v", resp)
		updateDevice(resp.Sid, resp.Model, resp.Data)
	}
}

func dataToJson(model string, data interface{}) interface{} {
	log.Printf("Data: %+v - type %+v", data, reflect.TypeOf(data))
	var retval interface{}
	switch model {
	case "motion":
		retval = MotionData{}
	case "sensor_ht":
		retval = Sensor_htData{}
	case "switch":
		retval = SwitchData{}
	case "magnet":
		retval = MagnetData{}
	case "gateway":
		retval = GatewayData{}
	}
	err := json.Unmarshal([]byte(data.(string)), &retval)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Json: %+v - type %+v", retval, reflect.TypeOf(retval))
	return retval
}

func updateDevice(sid string, model string, data interface{}) {
	if _, ok := devices[sid]; !ok {
		devices[sid] = object{
			"sid":   sid,
			"sodel": model,
		}
	}
	val := dataToJson(model, data)
	dv := devices[sid]
	log.Printf("Before Update: %+v", dv)
	for k, v := range val.(map[string]interface{}) {
		log.Printf("k: %+v - v: %+v", k, v)
		dv[k] = v
	}
	log.Printf("After Update: %+v", devices[sid])

	/*
		switch model {
		case "sensor_ht":
			log.Printf("Before Update: %+v", devices[sid])
			devices[sid].Voltage = val.(Sensor_htData).Voltage
			devices[sid].Temperature = val.(Sensor_htData).Temperature
			devices[sid].Humidity = val.(Sensor_htData).Humidity
			log.Printf("After Update: %+v", devices[sid])
		}
	*/
}

func serveUDP(a string) {
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		log.Fatal(err)
	}
	conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Panic(err)
	}
}

func serveMulticastUDP(a string) {
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		log.Fatal(err)
	}
	mconn, err = net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Panic(err)
	}
}

func loopReadMulticast(mconn *net.UDPConn, msgHandler func(resp *Response)) {
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
	serveUDP(gatewayIp + ":" + gatewayPort)

	go loopReadUdp(conn, msgHandler)
	go loopReadMulticast(mconn, msgHandler)

	gateways.Addr, err = net.ResolveUDPAddr("udp", multicastIp+":4321") //+multicastPort)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("sending whois to %+v from %+v\n", gateways, conn)
	gateways.sendMessage("whois", "")

	for {
		if len(devices) != 0 {
			for k, v := range devices {
				gateways.sendMessage("read", v["sid"].(string))
				switch v["model"] {
				case "sensor_ht":
					log.Printf("k: %s - v: %+v", k, v)
				case "motion":
				case "magnet":
				case "switch":
				}
			}
		}
		time.Sleep(time.Minute * 5)
	}
}
