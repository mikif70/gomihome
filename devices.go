// devices
package main

type DataIdList []string

type GatewayData struct {
	Rgb          int    `json:"rgb,omitempty"`
	Illumination int    `json:"illumination,omitempty"`
	ProtoVersion string `json:"proto_version,omitempty"`
}

type MotionData struct {
	Voltage int    `json:"voltage,omitempty"`
	Status  string `json:"status,omitempty"`
}

type MagnetData struct {
	Voltage int    `json:"voltage,omitempty"`
	Status  string `json:"status,omitempty"`
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
