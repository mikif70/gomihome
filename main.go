// main
package main

import (
	//	"encoding/hex"
	"fmt"
	"log"
	"net"
	"time"
)

type (
	Gateway struct {
		Addr  *net.UDPAddr
		Sid   string
		Token string
	}

	Request struct {
		Cmd string `json:"cmd"`
		Sid string `json:"sid,omitempty"`
	}
)

const (
	multicastIp     = "224.0.0.50"
	multicastPort   = "9898"
	maxDatagramSize = 8192
)

var (
	conn     *net.UDPConn
	gateways map[string]Gateway
)

func main() {
	ping(multicastIp + ":" + multicastPort)
}

func ping(a string) {
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		fmt.Printf("Err: %+v\n", err)
		log.Fatal(err)
	}
	c, err := net.DialUDP("udp", nil, addr)
	for {
		c.Write([]byte("hello, world\n"))
		time.Sleep(1 * time.Second)
	}
}
