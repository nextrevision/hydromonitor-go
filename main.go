package main

import (
	"flag"
	"time"

	"runtime"

	log "github.com/Sirupsen/logrus"
	"github.com/currantlabs/ble"
	"github.com/currantlabs/ble/darwin"
	"github.com/currantlabs/ble/linux"
)

var (
	debug          = flag.Bool("debug", false, "enable debug logging")
	scanInterval   = flag.Duration("scan-interval", 5*time.Minute, "time in minutes between scans for devices")
	pollInterval   = flag.Duration("poll-interval", 60*time.Minute, "time in minutes between refreshing device metrics")
	connectTimeout = flag.Duration("timeout", 15*time.Second, "timeout in seconds when connecting to devices")
	database       = flag.String("database", "hydromonitor.sql", "path to create database")
)

func main() {
	flag.Parse()

	log.SetLevel(log.InfoLevel)
	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	datastore := NewDatastore(*database)
	defer datastore.Close()

	var device ble.Device
	var err error
	switch runtime.GOOS {
	case "darwin":
		device, err = darwin.NewDevice()
	case "linux":
		device, err = linux.NewDevice()
	default:
		log.Fatalf("Unsupported OS: %s", runtime.GOOS)
	}
	if err != nil {
		log.Fatalf("Error creating device : %s", err)
	}
	ble.SetDefaultDevice(device)

	state := NewState(datastore, *connectTimeout)

	// Scan for specified duration, or until interrupted by user.
	go state.Scan(*scanInterval)
	go state.Poll(*pollInterval)

	api := NewAPI(datastore, state)
	api.Start()
}
