// multicast
package main

import (
	"encoding/json"
	"log"
	"net"
	"time"
)

var (
	mchan MChan
)

type MPack struct {
	packet    []byte
	timestamp time.Time
}

type MChan chan MPack

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

	mchan = make(chan MPack)

	return multicast
}

func (mu *Multicast) DiscoverGateway(gw *Udp) {
	wg.Add(1)
	mu.Gateway = gw
	mu.discover = true
	mu.resolveAddr()
	mu.dial()
	go mu.read()
	go mu.unarshallPacket()
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

	log.Printf("start multicast reading....")
	for mu.running {
		b := make([]byte, MaxDatagramSize)
		n, _, err := mu.conn.ReadFrom(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}

		nb := make([]byte, n)
		copy(nb, b[:n])

		mchan <- MPack{
			packet:    nb,
			timestamp: time.Now(),
		}
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
	log.Printf("Msg: %+v - Addr: %+v", string(req), mu.waddr)
	n, err = mu.conn.WriteTo([]byte(req), mu.waddr)
	if err != nil {
		log.Printf("Write error: %+v", err)
	}
	log.Printf("Wrote %d bytes", n)
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
		mu.unmarshallData(resp)
	case "report":
		//		log.Printf("REPORT: %+v", resp)
		mu.unmarshallData(resp)
	default:
		log.Printf("DEFAULT: %+v", resp)
	}
}

func (mu *Multicast) unarshallPacket() {

	var lastTimestamp time.Time
	var lastSid string
	var lastCmd string

	for mu.running {

		b := <-mchan

		resp := Response{}
		err := json.Unmarshal(b.packet, &resp)
		if err != nil {
			log.Printf("JSON Err: %+v - %+v", err, b.packet)
			continue
		}

		if b.timestamp.Equal(lastTimestamp) && resp.Cmd == lastCmd && resp.Sid == lastSid {
			continue
		}

		lastTimestamp = b.timestamp
		lastCmd = resp.Cmd
		lastSid = resp.Sid

		mu.msgHandler(&resp)
	}
}

func (mu *Multicast) unmarshallData(resp *Response) {
	switch resp.Model {
	case "motion":
		dt := MotionData{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		log.Printf("Motion (%s): %+v", resp.Cmd, dt)
	case "magnet":
		dt := MagnetData{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		log.Printf("Magnet (%s): %+v", resp.Cmd, dt)
	case "sensor_ht":
		dt := Sensor_htData{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		log.Printf("Sensor_HT (%s): %+v", resp.Cmd, dt)
	case "switch":
		dt := SwitchData{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		log.Printf("Switch (%s): %+v", resp.Cmd, dt)
	case "gateway":
		dt := GatewayData{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		log.Printf("Gateway (%s): %+v", resp.Cmd, dt)
	default:
		log.Printf("Model not defined: %s", resp.Model)
	}
}
