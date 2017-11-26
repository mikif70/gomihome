// gateway
package main

import (
	"encoding/json"
	"log"
	"net"
)

type Gateway struct {
	IP      string
	Port    string
	Sid     string
	Running bool
	Conn    *net.UDPConn
	Addr    *net.UDPAddr
}

func (gw *Gateway) serveUDP() {
	var err error

	gw.Addr, err = net.ResolveUDPAddr("udp", multicastIp+":"+gw.Port)
	if err != nil {
		log.Fatal(err)
	}
	conn, err = net.ListenMulticastUDP("udp", nil, gw.Addr)
	if err != nil {
		log.Panic(err)
	}
	gw.Conn = conn
	gw.Running = true
	go gw.loopReadUDP(gw.msgHandler)
}

func (gw *Gateway) loopReadUDP(msgHandler func(resp *Response)) {
	conn.SetReadBuffer(maxDatagramSize)

	for gw.Running {
		b := make([]byte, maxDatagramSize)
		n, _, err := gw.Conn.ReadFromUDP(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}

		resp := Response{}
		err = json.Unmarshal(b[:n], &resp)
		if err != nil {
			log.Printf("JSON Err: %+v", err)
			continue
		}
		msgHandler(&resp)
	}
	wg.Done()
}

func (gw *Gateway) msgHandler(resp *Response) {
	switch resp.Cmd {
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

func (gw *Gateway) sendMessage(msg string, sid string) {

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
	log.Printf("Msg: %+v - Addr: %+v", string(req), gw.Addr)
	conn.WriteMsgUDP([]byte(req), nil, gw.Addr)
}

func (gw *Gateway) discoverDevs() {
	gw.sendMessage("get_id_list", "")
}
