package main

import (
	"context"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/currantlabs/ble"
	"github.com/pkg/errors"
)

// State represents the current in-memory state of discovered clients
type State struct {
	tilts          map[string]*TiltClient
	datastore      *Datastore
	lockChan       chan int
	connectTimeout time.Duration
}

// NewState should only be called once to return an initial device state
func NewState(datastore *Datastore, connectTimeout time.Duration) *State {
	return &State{
		tilts:          make(map[string]*TiltClient),
		datastore:      datastore,
		connectTimeout: connectTimeout,
		lockChan:       make(chan int, 1),
	}
}

// Scan ...
func (s *State) Scan(interval time.Duration) {
	for {
		log.Debugf("[scan] Waiting for lock...")
		s.lockChan <- 1

		log.Infof("[scan] Scanning for new tilts...")
		deviceQueue := map[string]bool{}
		ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), s.connectTimeout))
		if err := ble.Scan(ctx, false, func(a ble.Advertisement) {
			log.Debugf("[scan] Found tilt: %s", a.Address())
			tiltClient, err := NewTiltClient(a.Address(), s.connectTimeout)
			if err != nil {
				log.Errorf("[scan] Error connecting to tilt %s: %s", a.Address(), err)
			} else {
				go s.addTilt(tiltClient)
			}
		}, func(a ble.Advertisement) bool {
			// Exclude devices in the device queue
			// TODO: if ok for presence
			for d := range deviceQueue {
				if d == a.Address().String() {
					log.Debug("[scan] Ignoring device already in queue")
					return false
				}
			}

			// TODO: if ok for presence
			for id := range s.tilts {
				if id == a.Address().String() {
					log.Debug("[scan] Ignoring device already discovered")
					return false
				}
			}

			// Only include devices named Tilt
			if a.LocalName() == "Tilt" {
				log.Debug("[scan] Queuing scan of tilt...")
				deviceQueue[a.Address().String()] = true
				return true
			}
			return false
		}); err != nil {
			if errors.Cause(err) == context.DeadlineExceeded || err == nil {
				log.Debug("[scan] Finished")
			} else if errors.Cause(err) == context.Canceled {
				// TODO: notify poll to stop?
				break
			} else {
				log.Fatalf("[scan] Could not start: %s", err)
			}
		}

		log.Debugf("[scan] Releasing lock...")
		<-s.lockChan

		log.Infof("[scan] Waiting %s before next scan...", interval)
		time.Sleep(interval)
	}
}

// Poll ...
func (s *State) Poll(interval time.Duration) {
	for {
		log.Debugf("[poll] Waiting for lock...")
		s.lockChan <- 1

		log.Debugf("[poll] Starting poll for %d tilts...", len(s.tilts))
		for id := range s.tilts {
			log.Debugf("[poll] Refreshing %s...", id)
			err := s.RefreshTilt(id)
			if err != nil {
				log.Errorf("[poll] Error refreshing metrics for tilt: %s", id)
				log.Error(err)
				s.recordError(id, err)
			}
		}

		log.Debugf("[poll] Releasing lock...")
		<-s.lockChan

		log.Debugf("[poll] Waiting %s before next polling run...", interval)
		time.Sleep(interval)
	}
}

// RefreshTilt ...
func (s *State) RefreshTilt(tiltID string) error {
	// Verify the device actually exists in the state
	if !s.tiltInState(tiltID) {
		return fmt.Errorf("No such tilt: %s", tiltID)
	}

	metric, err := s.tilts[tiltID].RefreshMetrics(s.connectTimeout)
	if err != nil {
		return err
	}

	log.Debugf("Creating metric: %+v", metric)
	if err = s.datastore.CreateMetric(metric); err != nil {
		return fmt.Errorf("Error storing device metric: %s", err)
	}

	return nil
}

func (s *State) addTilt(tilt *TiltClient) {
	log.Debugf("[state] Adding tilt to database: %s", tilt.Address)
	if err := s.datastore.CreateOrUpdateDevice(Device{ID: tilt.Address.String(), Color: tilt.Color}); err != nil {
		log.Errorf("[state] Error storing tilt information: %s", err)
	}

	log.Debugf("[state] Adding tilt to state: %s", tilt.Address)
	s.tilts[tilt.Address.String()] = tilt
	if err := s.RefreshTilt(tilt.Address.String()); err != nil {
		log.Errorf("[state] Error refreshing tilt metrics: %s", err)
	}
}

func (s *State) recordError(tiltID string, e error) {
	if err := s.datastore.SetDeviceError(tiltID, e.Error()); err != nil {
		log.Errorf("Error setting tilt error: %s", err)
	}
}

func (s *State) tiltInState(tiltID string) bool {
	for id := range s.tilts {
		if id == tiltID {
			return true
		}
	}
	return false
}
