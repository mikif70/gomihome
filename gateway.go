// gateway
package main

import (
	"encoding/json"
	"log"
	"net"
)

type Gateway struct {
	IP              string
	Port            string
	MulticastIP     string
	MulticastPort   string
	MaxDatagramSize int
	sid             string
	running         bool
	conn            *net.UDPConn
	addr            *net.UDPAddr
	multicastRun    bool
	multicastAddr   *net.UDPAddr
	multicastConn   *net.UDPConn
}

var (
	gateway = &Gateway{
		MulticastIP:     "224.0.0.50",
		MulticastPort:   "4321",
		MaxDatagramSize: 1024,
	}
)

func (gw *Gateway) DiscoverGateway() {
	wg.Add(1)
	gw.resolveMulticastAddr()
	gw.dialMulticast()
	go gw.ReadMulticast()
	gw.writeMulticast("whois", "")
}

func (gw *Gateway) discoverDevs() {
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

func (gw *Gateway) resolveMulticastAddr() {
	var err error

	gw.multicastAddr, err = net.ResolveUDPAddr("udp", gw.MulticastIP+":"+gw.MulticastPort)
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

func (gw *Gateway) dialMulticast() {
	var err error

	gw.multicastConn, err = net.ListenMulticastUDP("udp", nil, gw.multicastAddr)
	if err != nil {
		log.Panic(err)
	}
	gw.multicastRun = true
}

func (gw *Gateway) readUDP() {

	for gw.running {

		log.Printf("Reading UDP: %+v", gw.addr)

		//	conn.SetReadBuffer(maxDatagramSize)

		b := make([]byte, gw.MaxDatagramSize)
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

func (gw *Gateway) ReadMulticast() {

	for gw.multicastRun {
		b := make([]byte, gw.MaxDatagramSize)
		n, _, err := gw.multicastConn.ReadFrom(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}

		resp := Response{}
		err = json.Unmarshal(b[:n], &resp)
		if err != nil {
			log.Printf("JSON Err: %+v", err)
			continue
		}
		gw.msgHandler(&resp)
	}
	gw.multicastConn.Close()
}

func (gw *Gateway) msgHandler(resp *Response) {
	switch resp.Cmd {
	case "iam":
		log.Printf("IAM: %+v", resp)
		gateway.sid = resp.Sid
		gateway.IP = resp.IP
		gateway.Port = resp.Port
		log.Printf("Gateway: %+v", gateway)
		gw.multicastRun = false
		wg.Done()
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

func (gw *Gateway) writeMulticast(msg string, sid string) {

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
	gw.multicastConn.Write([]byte(req))
}
