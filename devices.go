// devices
package main

import (
	"encoding/json"
	"log"
	"strconv"
	"time"
)

/*
	Gateway:
		rgb:
			<int>
		illumination:
			<int>


	Switch:
		voltage:
			<int>
		status:
			"click"
			"double_click"
			"long_click_press"
			"long_click_release"

	Motion:
		voltage:
			<int>
		status:
			"motion"
		no_motion:
			<sec>

	Sensor_ht:
		voltage:
			<int>
		temperature:
			<int>
		humidity:
			<int>

*/

type Device struct {
	Name  string `json:"name"`
	Model string `json:"model"`
	Sid   string `json:"sid"`
}

type DataIdList []string

type InfluxDevice struct {
	Cmd          string
	Model        string
	Sid          string
	Voltage      int
	Status       string
	Illumination int
	Rgb          int
	ProtoVersion string
	Motion       bool
	NoMotion     int
	Open         bool
	Close        bool
	NoClose      int
	Temperature  int
	Humidity     int
	Timestamp    time.Time
}

type GatewayData struct {
	Rgb          int    `json:"rgb,omitempty"`
	Illumination int    `json:"illumination,omitempty"`
	ProtoVersion string `json:"proto_version,omitempty"`
}

type MotionData struct {
	Voltage  int    `json:"voltage,omitempty"`
	Status   string `json:"status,omitempty"`
	NoMotion string `json:"no_motion,omitempty"`
}

type MagnetData struct {
	Voltage int    `json:"voltage,omitempty"`
	Status  string `json:"status,omitempty"`
	NoClose string `json:"no_close,omitempty"`
}

type SwitchData struct {
	Voltage int    `json:"voltage,omitempty"`
	Status  string `json:"status,omitempty"`
}

type Sensor_htData struct {
	Voltage     int    `json:"voltage,omitempty"`
	Temperature string `json:"temperature,omitempty"`
	Humidity    string `json:"humidity,omitempty"`
}

func unmarshallData(resp *Response) {
	indevs := &InfluxDevice{
		Cmd:       resp.Cmd,
		Model:     resp.Model,
		Sid:       resp.Sid,
		Timestamp: time.Now(),
	}

	switch resp.Model {
	case "motion":
		dt := MotionData{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		indevs.Voltage = dt.Voltage
		indevs.Status = dt.Status
		switch dt.Status {
		case "motion":
			indevs.Motion = true
			indevs.NoMotion = 0
		case "no_motion":
			indevs.Motion = false
			indevs.NoMotion, _ = strconv.Atoi(dt.NoMotion)
		}
	case "magnet":
		dt := MagnetData{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		indevs.Voltage = dt.Voltage
		indevs.Status = dt.Status
		switch dt.Status {
		case "open":
			indevs.Close = false
			indevs.Open = true
			indevs.NoClose = 0
		case "close":
			indevs.Close = true
			indevs.Open = false
			indevs.NoClose = 0
		case "no_close":
			indevs.NoClose, _ = strconv.Atoi(dt.NoClose)
			indevs.Close = false
			indevs.Open = true
		}
	case "sensor_ht":
		dt := Sensor_htData{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		indevs.Voltage = dt.Voltage
		indevs.Temperature, _ = strconv.Atoi(dt.Temperature)
		indevs.Humidity, _ = strconv.Atoi(dt.Humidity)
	case "switch":
		dt := SwitchData{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		indevs.Voltage = dt.Voltage
		indevs.Status = dt.Status
	case "gateway":
		dt := GatewayData{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		indevs.Illumination = dt.Illumination
		indevs.Rgb = dt.Rgb
		indevs.ProtoVersion = dt.ProtoVersion
	default:
		log.Printf("Model not defined: %s", resp.Model)
	}

	log.Printf("%s: %s", resp.Model, resp.Cmd)

	if (resp.Cmd == "report" && resp.Model != "sensor_ht") || (resp.Cmd == "heartbeat" && resp.Model != "gateway") || (resp.Cmd == "read_ack" && (resp.Model == "sensor_ht" || resp.Model == "gateway")) {
		writeStats(indevs)
		if opts.Verbose || opts.Debug {
			log.Printf("%s Data: %+v", resp.Model, resp.Data)
		}
		if opts.Debug {
			log.Printf("%s InfluxData: %+v", resp.Model, indevs)
		}
	}
}
