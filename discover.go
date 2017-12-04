// multicast
package main

/*
import (
	"encoding/json"
	"log"
	"net"
)

var (
	multicast bool
)

type Multicast struct {
	Addr    *net.UDPAddr
	Conn    *net.UDPConn
	Running bool
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
	log.Printf("Msg: %+v - Addr: %+v", string(req), addr)
	conn.WriteTo([]byte(req), addr)
}

func discoverGateway(a string) {
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		log.Fatal(err)
	}
	conn, err = net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Panic(err)
	}
	go loopReadMulticast(conn, msgHandler)
	log.Printf("sending whois to %+v from %+v\n", addr, conn)
	sendMessage(addr, "whois", "")
}

func loopReadMulticast(conn net.PacketConn, msgHandler func(resp *Response)) {
	//	conn.SetReadBuffer(maxDatagramSize)

	for multicast {
		b := make([]byte, maxDatagramSize)
		n, _, err := conn.ReadFrom(b)
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
	conn.Close()
}

func msgHandler(resp *Response) {
	switch resp.Cmd {
	case "iam":
		log.Printf("IAM: %+v", resp)
		gateway.Sid = resp.Sid
		gateway.IP = resp.IP
		gateway.Port = resp.Port
		log.Printf("Gateway: %+v", gateway)
		multicast = false
		wg.Done()
		//	case "heartbeat":
		//		log.Printf("HEARTBEAT: %+v", resp)
		//	case "get_id_list_ack":
		//		log.Printf("Get ACK: %+v\n", resp)
		//	case "read_ack":
		//		log.Printf("Read ACK: %+v", resp)
	default:
		log.Printf("DEFAULT: %+v", resp)
	}
}
*/