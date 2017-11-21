package main

import (
	//	"encoding/hex"
	"encoding/json"
	"log"
	"net"
	"strings"
	"sync"
	//	"time"
)

type (
	Response struct {
		Cmd   string `json:"cmd"`
		Model string `json:"model"`
		Sid   string `json:"sid"`
		//ShortId int `json:"short_id,integer"`
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

	object map[string]interface{}
)

var (
	conn     *net.UDPConn
	gateways *Gateway
	wg       sync.WaitGroup
	whois    = 0
)

const (
	multicastIp   = "224.0.0.50"
	multicastPort = "9898"
	//multicastAddr         = "239.255.255.250:9898"
	maxDatagramSize = 8192
)

func GetStatus() string { //test
	return ""
}

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
	default:
		log.Printf("DEFAULT: %+v", resp)
	}
}

func serveMulticastUDP(a string) {
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		log.Fatal(err)
	}
	conn, err = net.ListenMulticastUDP("udp", nil, addr)
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
		//log.Println(n, "bytes read from", src)
		//log.Println(string(b[:n]))

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
	}
}
