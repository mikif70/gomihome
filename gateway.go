// gateway
package main

import (
	"encoding/json"
	"log"
	"net"
	//	"strings"
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
	gw.resolveAddr(gw.IP, gw.Port)
	gw.dial()
	go gw.read()
	gw.write("get_id_list", "")

}

func (gw *Gateway) resolveAddr(ip string, port string) {
	var err error

	gw.IP = ip
	gw.Port = port

	gw.addr, err = net.ResolveUDPAddr("udp", gw.IP+":"+gw.Port)
	if err != nil {
		log.Fatal(err)
	}
}

func (gw *Gateway) dial() {
	var err error

	gw.conn, err = net.DialUDP("udp", nil, gw.addr)
	if err != nil {
		log.Panic(err)
	}
	gw.running = true
}

func (gw *Gateway) read() {

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
		dt := DataIdList{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		//		retval := strings.Split(resp.Data.(string), ",")
		//		r := strings.NewReplacer("\"", "", "[", "", "]", "")
		for i := range dt.Id {
			//ns := r.Replace(retval[i])
			log.Printf("Data: %d - %s", i, dt.Id[i])
			gw.write("read", dt.Id[i])
		}
	case "read_ack":
		log.Printf("Read ACK: %+v", resp)
	default:
		log.Printf("DEFAULT: %+v", resp)
	}
}

func (gw *Gateway) write(msg string, sid string) {

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
