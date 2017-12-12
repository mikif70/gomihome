// main.go
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

const (
	_Version = "v1.1.0"
)

var (
	wg sync.WaitGroup
)

func main() {

	flag.Usage = usage
	flag.Parse()

	if opts.Version {
		fmt.Printf("%s: %s\n", os.Args[0], _Version)
		os.Exit(0)
	}

	f, err := os.OpenFile(opts.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	if opts.Debug {
		log.Printf("Opts: %+v", opts)
	}

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
