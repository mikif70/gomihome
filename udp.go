// udp
package main

import (
	"encoding/json"
	"log"
	"net"
	"time"
	//	"strings"
)

type Udp struct {
	IP      string
	Port    string
	sid     string
	running bool
	conn    *net.UDPConn
	addr    *net.UDPAddr
}

type Devices []Device

var (
	devices = make([]Device, 0)
	ticker  = time.NewTicker(1 * time.Minute)
	numdevs = 0
)

func newUdp() *Udp {
	udp := &Udp{}

	return udp
}

func (gw *Udp) DiscoverDevs() {
	wg.Add(1)
	gw.resolveAddr(gw.IP, gw.Port)
	gw.dial()
	go gw.read()
	gw.write("get_id_list", "")

}

func (gw *Udp) resolveAddr(ip string, port string) {
	var err error

	gw.IP = ip
	gw.Port = port

	gw.addr, err = net.ResolveUDPAddr("udp", gw.IP+":"+gw.Port)
	if err != nil {
		log.Fatal(err)
	}
}

func (gw *Udp) dial() {
	var err error

	gw.conn, err = net.DialUDP("udp", nil, gw.addr)
	if err != nil {
		log.Panic(err)
	}
	gw.running = true
}

func (gw *Udp) doReadDevs() {
	for t := range ticker.C {
		log.Printf("devs: %+v", devices)
		for d := range devices {
			gw.write("read", devices[d].Sid)
		}
		gw.write("read", gw.sid)
		log.Printf("Read: %+v", t)
	}
}

func (gw *Udp) read() {

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

func (gw *Udp) msgHandler(resp *Response) {
	switch resp.Cmd {
	case "get_id_list_ack":
		log.Printf("Get ACK: %+v\n", resp.Data)
		dt := DataIdList{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		numdevs = len(dt)
		for i := range dt {
			log.Printf("Data: %d - %s", i, dt[i])
			gw.write("read", dt[i])
		}
		gw.write("read", gw.sid)
		devices = append(devices, Device{
			Sid:   gw.sid,
			Model: "gateway",
		})
		go gw.doReadDevs()
	case "read_ack":
		//		log.Printf("Read ACK: %+v", resp)
		if numdevs != 0 {
			devices = append(devices, Device{
				Sid:   resp.Sid,
				Model: resp.Model,
			})
			numdevs--
		}
		gw.unmarshallData(resp)
	case "heartbeat":
		gw.unmarshallData(resp)
	default:
		log.Printf("DEFAULT: %+v", resp)
	}
}

func (gw *Udp) unmarshallData(resp *Response) {
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

func (gw *Udp) write(msg string, sid string) {

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
