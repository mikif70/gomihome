// multicast
package main

import (
	"encoding/json"
	"log"
	"net"
)

type Multicast struct {
	IP              string
	Port            string
	run             bool
	addr            *net.UDPAddr
	conn            *net.UDPConn
	MaxDatagramSize int
	Gateway         *Gateway
}

func newMulticast() *Multicast {
	multicast := &Multicast{
		IP:              "224.0.0.50",
		Port:            "4321",
		MaxDatagramSize: 1024,
	}

	return multicast
}

func (mu *Multicast) DiscoverGateway(gw *Gateway) {
	wg.Add(1)
	mu.resolveAddr()
	mu.dial()
	go mu.read()
	mu.write("whois", "")
}

func (mu *Multicast) resolveAddr() {
	var err error

	mu.addr, err = net.ResolveUDPAddr("udp", mu.IP+":"+mu.Port)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("multicast addr: %+v", mu.addr)
}

func (mu *Multicast) dial() {
	var err error

	mu.conn, err = net.ListenMulticastUDP("udp", nil, mu.addr)
	if err != nil {
		log.Panic(err)
	}
	mu.run = true
	log.Printf("multicast conn: %+v", mu.conn)
}

func (mu *Multicast) read() {

	log.Printf("start multicast reading....")
	for mu.run {
		b := make([]byte, mu.MaxDatagramSize)
		n, _, err := mu.conn.ReadFrom(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}

		resp := Response{}
		err = json.Unmarshal(b[:n], &resp)
		if err != nil {
			log.Printf("JSON Err: %+v", err)
			continue
		}
		mu.msgHandler(&resp)
	}
	mu.conn.Close()
}

func (mu *Multicast) write(msg string, sid string) {

	var req []byte
	var err error
	var n int

	if sid != "" {
		req, err = json.Marshal(Request{Cmd: msg, Sid: sid})
	} else {
		req, err = json.Marshal(Request{Cmd: msg})
	}

	if err != nil {
		log.Printf("Marshall error: %+v", err)
	}
	log.Printf("Msg: %+v - Addr: %+v", string(req), mu.addr)
	n, err = mu.conn.WriteTo([]byte(req), mu.addr)
	if err != nil {
		log.Printf("Write error: %+v", err)
	}
	log.Printf("Wrote %d bytes", n)
}

func (mu *Multicast) msgHandler(resp *Response) {
	switch resp.Cmd {
	case "iam":
		log.Printf("IAM: %+v", resp)
		mu.Gateway.sid = resp.Sid
		mu.Gateway.IP = resp.IP
		mu.Gateway.Port = resp.Port
		log.Printf("mu.Gateway: %+v", mu.Gateway)
		//		gw.multicastRun = false
		wg.Done()
	default:
		log.Printf("DEFAULT: %+v", resp)
	}
}
