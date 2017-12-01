// protocol
package main

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

type GatewayData struct {
	Ip           string `json:"ip"`
	Port         string `json:"port"`
	Rgb          int    `json:"rgb,omitempty"`
	Illumination int    `json:"illumination,omitempty"`
}
