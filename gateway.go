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
	sid     string
	running bool
	conn    *net.UDPConn
	addr    *net.UDPAddr
}

func newGateway() *Gateway {
	gateway := &Gateway{}

	return gateway
}

func (gw *Gateway) DiscoverDevs() {
	wg.Add(1)
	gw.resolveUDPAddr(gw.IP, gw.Port)
	gw.dialUDP()
	go gw.readUDP()
	gw.writeUdp("get_id_list", "")

}

func (gw *Gateway) resolveUDPAddr(ip string, port string) {
	var err error

	gw.IP = ip
	gw.Port = port

	gw.addr, err = net.ResolveUDPAddr("udp", gw.IP+":"+gw.Port)
	if err != nil {
		log.Fatal(err)
	}
}

func (gw *Gateway) dialUDP() {
	var err error

	gw.conn, err = net.DialUDP("udp", nil, gw.addr)
	if err != nil {
		log.Panic(err)
	}
	gw.running = true
}

func (gw *Gateway) readUDP() {

	for gw.running {

		log.Printf("Reading UDP: %+v", gw.addr)

		//	conn.SetReadBuffer(maxDatagramSize)

		b := make([]byte, MaxDatagramSize)
		n, err := gw.conn.Read(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}

		log.Printf("Read from UDP: %d bytes", n)

		resp := Response{}
		err = json.Unmarshal(b[:n], &resp)
		if err != nil {
			log.Printf("JSON Err: %+v", err)
			continue
		}
		gw.msgHandler(&resp)
	}
	gw.conn.Close()
}

func (gw *Gateway) msgHandler(resp *Response) {
	switch resp.Cmd {
	case "get_id_list_ack":
		log.Printf("Get ACK: %+v\n", resp.Data)
	case "read_ack":
		log.Printf("Read ACK: %+v", resp)
	default:
		log.Printf("DEFAULT: %+v", resp)
	}
}

func (gw *Gateway) writeUdp(msg string, sid string) {

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
	log.Printf("Msg: %+v - Addr: %+v", string(req), gw.conn)
	gw.conn.Write([]byte(req))
}
