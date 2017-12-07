// options
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Options struct {
	LogFile     string
	CurrentPath string
	Exe         string
	Influxdb    string
	Debug       bool
	Verbose     bool
	Version     bool
	TimeToTick  time.Duration
}

var (
	opts = Options{
		LogFile:    "xiaomi.log",
		TimeToTick: 10 * time.Minute,
	}
)

func usage() {
	fmt.Println(`Usage: ` + path.Base(os.Args[0]) + `
		-i <influxdb [ip:port]> 
		-t <check interval> 
		-l <logfile> 
		-v <version>
		-V <verbose>
		-D <debug>`)
	fmt.Println()
	os.Exit(0)
}

func init() {
	var err error
	opts.CurrentPath, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	opts.LogFile = path.Join(opts.CurrentPath, opts.LogFile)
	opts.Exe = path.Base(os.Args[0])

	flag.StringVar(&opts.Influxdb, "i", "127.0.0.1:8086", "influxdb server")
	flag.StringVar(&opts.LogFile, "l", opts.LogFile, "Logs filename")
	flag.BoolVar(&opts.Version, "v", false, "Version")
	flag.DurationVar(&opts.TimeToTick, "t", opts.TimeToTick, "Check interval")
	flag.BoolVar(&opts.Debug, "D", false, "Debug")
	flag.BoolVar(&opts.Verbose, "V", false, "Verbose")
}
