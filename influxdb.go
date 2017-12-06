// influxdb
package main

import (
	"fmt"
	"log"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
)

const (
	Infdb = "192.168.1.10:8086"
)

func writeStats(id *InfluxDevice) {
	if DEBUG {
		log.Printf("writing to influxdb server: %s", Infdb)
		log.Printf("Devs: %+v", id)
	}

	c, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:    fmt.Sprintf("http://%s", Infdb),
		Timeout: 2 * time.Second,
	})
	if err != nil {
		log.Printf("Error: %+v\n", err)
		return
	}
	defer c.Close()

	bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:  "xiaomi",
		Precision: "s",
	})
	if err != nil {
		log.Printf("Error: %+v\n", err)
		return
	}

	tags := map[string]string{"sid": id.Sid, "model": id.Model}
	fields := map[string]interface{}{
		"voltage":      id.Voltage,
		"status":       id.Status,
		"illumination": id.Illumination,
		"rgb":          id.Rgb,
		"nomotion":     id.NoMotion,
		"noclose":      id.NoClose,
		"temperature":  id.Temperature,
		"humidity":     id.Humidity,
		"open":         id.Open,
		"close":        id.Close,
		"motion":       id.Motion,
	}
	pt, err := influxdb.NewPoint("gateway", tags, fields, id.Timestamp)
	if err != nil {
		log.Printf("Error: %+v\n", err)
		return
	}

	bp.AddPoint(pt)

	// Write the batch
	if err := c.Write(bp); err != nil {
		log.Printf("Influxdb write: ", err)
	}
}
