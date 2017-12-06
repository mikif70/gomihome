// devices
package main

import (
	"encoding/json"
	"log"
	"time"
)

type Device struct {
	Name  string `json:"name"`
	Model string `json:"model"`
	Sid   string `json:"sid"`
}

type DataIdList []string

type InfluxDevice struct {
	Model        string
	Sid          string
	Voltage      int
	Status       string
	Illumination int
	Rgb          int
	NoMotion     int
	NoClose      int
	Temperature  string
	Humidity     string
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
	NoMotion int    `json:"no_motion,omitempty"`
}

type MagnetData struct {
	Voltage int    `json:"voltage,omitempty"`
	Status  string `json:"status,omitempty"`
	NoClose int    `json:"no_close,omitempty"`
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
		log.Printf("Motion (%s): %+v", resp.Cmd, dt)
		indevs.Voltage = dt.Voltage
		indevs.Status = dt.Status
		indevs.NoMotion = dt.NoMotion
	case "magnet":
		dt := MagnetData{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		log.Printf("Magnet (%s): %+v", resp.Cmd, dt)
		indevs.Voltage = dt.Voltage
		indevs.Status = dt.Status
		indevs.NoClose = dt.NoClose
	case "sensor_ht":
		dt := Sensor_htData{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		log.Printf("Sensor_HT (%s): %+v", resp.Cmd, dt)
		indevs.Voltage = dt.Voltage
		indevs.Temperature = dt.Temperature
		indevs.Humidity = dt.Humidity
	case "switch":
		dt := SwitchData{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		log.Printf("Switch (%s): %+v", resp.Cmd, dt)
		indevs.Voltage = dt.Voltage
		indevs.Status = dt.Status
	case "gateway":
		dt := GatewayData{}
		err := json.Unmarshal([]byte(resp.Data.(string)), &dt)
		if err != nil {
			log.Printf("JSON Data Err: %+v", err)
			return
		}
		log.Printf("Gateway (%s): %+v", resp.Cmd, dt)
		indevs.Illumination = dt.Illumination
		indevs.Rgb = dt.Rgb
	default:
		log.Printf("Model not defined: %s", resp.Model)
	}

	if (resp.Cmd == "heartbeat" && resp.Model != "gateway") || (resp.Cmd == "read_ack" && (resp.Model == "sensor_ht" || resp.Model == "gateway")) {
		writeStats(indevs)
	}
}
