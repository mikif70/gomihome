// influxdb
package main

import (
	"fmt"
	"log"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
)

func writeStats(id *InfluxDevice) {
	if opts.Verbose || opts.Debug {
		log.Printf("writing to influxdb server: %s", opts.Influxdb)
		log.Printf("Devs: %+v", id)
	}

	c, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:    fmt.Sprintf("http://%s", opts.Influxdb),
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

	tags := map[string]string{"sid": id.Sid, "model": id.Model, "cmd": id.Cmd}

	var fields map[string]interface{}

	switch id.Model {
	case "motion":
		fields = map[string]interface{}{
			"voltage":  id.Voltage,
			"status":   id.Status,
			"nomotion": id.NoMotion,
			"motion":   id.Motion,
		}
	case "magnet":
		fields = map[string]interface{}{
			"voltage": id.Voltage,
			"status":  id.Status,
			"noclose": id.NoClose,
			"open":    id.Open,
			"close":   id.Close,
		}
	case "sensor_ht":
		fields = map[string]interface{}{
			"voltage":     id.Voltage,
			"status":      id.Status,
			"temperature": id.Temperature,
			"humidity":    id.Humidity,
		}
	case "weather.v1":
		fields = map[string]interface{}{
			"voltage":     id.Voltage,
			"status":      id.Status,
			"temperature": id.Temperature,
			"humidity":    id.Humidity,
			"pressure":    id.Pressure,
		}
	case "switch":
		fields = map[string]interface{}{
			"voltage": id.Voltage,
			"status":  id.Status,
		}
	case "gateway":
		fields = map[string]interface{}{
			"illumination": id.Illumination,
			"rgb":          id.Rgb,
			"protoversion": id.ProtoVersion,
		}
	}
	pt, err := influxdb.NewPoint(id.Model, tags, fields, id.Timestamp)
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
