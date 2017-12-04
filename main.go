// newmain.go
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

package main

import (
	//	"encoding/json"
	"io"
	"log"
	"os"
	//	"strings"
	"sync"
	//	"time"
)

var (
	wg sync.WaitGroup
)

func main() {
	f, err := os.OpenFile("xiaomi.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	udp := newUdp()
	multicast := newMulticast()

	log.Println("Starting Discover gateway...")
	multicast.DiscoverGateway(udp)
	wg.Wait()
	log.Printf("Gateway: %+v", udp)

	log.Println("Starting Discover devices...")
	udp.DiscoverDevs()
	wg.Wait()
}
