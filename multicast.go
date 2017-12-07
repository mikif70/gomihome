// multicast
package main

import (
	"encoding/json"
	"log"
	"net"
	"time"
)

type Multicast struct {
	running  bool
	discover bool
	raddr    *net.UDPAddr
	waddr    *net.UDPAddr
	conn     *net.UDPConn
	Gateway  *Udp
}

func newMulticast() *Multicast {
	multicast := &Multicast{}

	return multicast
}

func (mu *Multicast) DiscoverGateway(gw *Udp) {
	wg.Add(1)
	mu.Gateway = gw
	mu.discover = true
	mu.resolveAddr()
	mu.dial()
	go mu.read()
	mu.write("whois", "")
}

func (mu *Multicast) resolveAddr() {
	var err error

	mu.raddr, err = net.ResolveUDPAddr("udp", MulticastIP+":"+MulticastRPort)
	if err != nil {
		log.Fatal(err)
	}
	mu.waddr, err = net.ResolveUDPAddr("udp", MulticastIP+":"+MulticastWPort)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("multicast addr: %+v - %+v", mu.raddr, mu.waddr)
}

func (mu *Multicast) dial() {
	var err error

	mu.conn, err = net.ListenMulticastUDP("udp", nil, mu.raddr)
	if err != nil {
		log.Panic(err)
	}
	mu.running = true
	log.Printf("multicast conn: %+v", mu.conn)
}

func (mu *Multicast) read() {

	var lastTimestamp, timestamp int64
	var lastSid string
	var lastCmd string

	log.Printf("start multicast reading....")
	for mu.running {
		b := make([]byte, MaxDatagramSize)
		n, _, err := mu.conn.ReadFrom(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}

		timestamp = time.Now().Unix()

		resp := Response{}
		err = json.Unmarshal(b[:n], &resp)
		if err != nil {
			log.Printf("JSON Err: %+v - %+v", err, b[:n])
			continue
		}

		if timestamp == lastTimestamp && resp.Cmd == lastCmd && resp.Sid == lastSid {
			//			log.Printf("new: %+v - %s - %s", timestamp, resp.Cmd, resp.Sid)
			//			log.Printf("old: %+v - %s - %s", lastTimestamp, lastCmd, lastSid)
			continue
		}

		lastTimestamp = timestamp
		lastCmd = resp.Cmd
		lastSid = resp.Sid

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
	if opts.Debug {
		log.Printf("Msg: %+v - Addr: %+v", string(req), mu.waddr)
	}
	n, err = mu.conn.WriteTo([]byte(req), mu.waddr)
	if err != nil {
		log.Printf("Write error: %+v", err)
	}
	if opts.Debug {
		log.Printf("Wrote %d bytes", n)
	}
}

func (mu *Multicast) msgHandler(resp *Response) {
	switch resp.Cmd {
	case "iam":
		if mu.discover {
			log.Printf("IAM: %+v", resp)
			mu.Gateway.sid = resp.Sid
			mu.Gateway.IP = resp.IP
			mu.Gateway.Port = resp.Port
			log.Printf("mu.Gateway: %+v", mu.Gateway)
			mu.discover = false
			wg.Done()
		}
	case "heartbeat":
		//		log.Printf("Heartbeat: %s - %s", resp.Model, resp.Sid)
		unmarshallData(resp)
	case "report":
		//		log.Printf("REPORT: %+v", resp)
		unmarshallData(resp)
	default:
		if opts.Debug {
			log.Printf("DEFAULT: %+v", resp)
		}
	}
}
