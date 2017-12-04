// protocol
package main

const (
	MulticastIP     = "224.0.0.50"
	MulticastWPort  = "4321"
	MulticastRPort  = "9898"
	MaxDatagramSize = 1024
)

type Response struct {
	Cmd   string      `json:"cmd"`
	Model string      `json:"model"`
	Sid   string      `json:"sid"`
	Token string      `json:"token,omitempty"`
	IP    string      `json:"ip,omitempty"`
	Port  string      `json:"port,omitempty"`
	Data  interface{} `json:"data"`
}

type Request struct {
	Cmd string `json:"cmd"`
	Sid string `json:"sid,omitempty"`
}
