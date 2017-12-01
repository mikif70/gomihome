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

func (gw *Gateway) dialUDP() {
	var err error

	gw.Addr, err = net.ResolveUDPAddr("udp", gw.IP+":"+gw.Port)
	if err != nil {
		log.Fatal(err)
	}
	gw.Conn, err = net.DialUDP("udp", nil, gw.Addr)
	if err != nil {
		log.Panic(err)
	}
	gw.Running = true
}

func (gw *Gateway) readUDP() (resp *Response, err error) {

	log.Printf("Reading UDP: %+v", gw.Addr)

	//	conn.SetReadBuffer(maxDatagramSize)

	b := make([]byte, maxDatagramSize)
	n, err := gw.Conn.Read(b)
	if err != nil {
		log.Fatal("ReadFromUDP failed:", err)
	}

	log.Printf("Read from UDP: %d bytes", n)

	resp = &Response{}
	err = json.Unmarshal(b[:n], &resp)
	if err != nil {
		log.Printf("JSON Err: %+v", err)
		return nil, err
	}

	return resp, nil

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
	log.Printf("Msg: %+v - Addr: %+v", string(req), gw.Conn)
	gw.Conn.Write([]byte(req))
}

func (gw *Gateway) discoverDevs() {
	gw.sendMessage("get_id_list", "")
	resp, err := gw.readUDP()
	if err != nil {
		log.Printf("Read err: %+v", err)
		return
	}
	gw.msgHandler(resp)
}
