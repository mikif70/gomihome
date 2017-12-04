// newmain.go
/*
	Gateway:
		rgb:
			<int>
		illumination:
			<int>


	Switch:
		status:
			"click"
			"double_click"
			"long_click_press"
			"long_click_release"

	Motion:
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

	gateway := newGW()
	multicast := newMulticast()

	log.Println("Starting Discover gateway...")
	multicast.DiscoverGateway(gateway)
	wg.Wait()
	log.Printf("Gateway: %+v", gateway)

	log.Println("Starting Discover devices...")
	gateway.DiscoverDevs()
	wg.Wait()
}
