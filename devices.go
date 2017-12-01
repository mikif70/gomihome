// devices
package main

type MotionData struct {
	Voltage int    `json:"voltage"`
	Status  string `json:"status`
}

type MagnetData struct {
	Voltage int    `json:"voltage"`
	Status  string `json:"status`
}

type SwitchData struct {
	Voltage int    `json:"voltage"`
	Status  string `json:"status"`
}

type Sensor_htData struct {
	Voltage     int    `json:"voltage"`
	Temperature string `json:"temperature"`
	Humidity    string `json:"humidity"`
}
