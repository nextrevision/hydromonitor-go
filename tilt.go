package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"encoding/hex"

	log "github.com/Sirupsen/logrus"
	"github.com/currantlabs/ble"
)

const (
	BatteryID     = "2a19"
	TemperatureID = "a495ff22c5b14b44b5121370f02d74de"
	GravityID     = "a495ff23c5b14b44b5121370f02d74de"
	ColorID       = "a495ff25c5b14b44b5121370f02d74de"
	ColorRed      = "10bb"
	ColorGreen    = "20bb"
	ColorBlack    = "30bb"
	ColorPurple   = "40bb"
	ColorOrange   = "50bb"
	ColorBlue     = "60bb"
	ColorYellow   = "70bb"
	ColorPink     = "80bb"
)

var (
	batteryChar     *ble.Characteristic
	temperatureChar *ble.Characteristic
	gravityChar     *ble.Characteristic
	colorChar       *ble.Characteristic
)

type TiltClient struct {
	Address ble.Addr
	Color   string
	Errors  int
}

func init() {
	batteryUUID, err := ble.Parse(BatteryID)
	if err != nil {
		log.Fatal("Could not parse BatteryID")
	}
	batteryChar = ble.NewCharacteristic(batteryUUID)

	temperatureUUID, err := ble.Parse(TemperatureID)
	if err != nil {
		log.Fatal("Could not parse TemperatureID")
	}
	temperatureChar = ble.NewCharacteristic(temperatureUUID)

	gravityUUID, err := ble.Parse(GravityID)
	if err != nil {
		log.Fatal("Could not parse GravityID")
	}
	gravityChar = ble.NewCharacteristic(gravityUUID)

	colorUUID, err := ble.Parse(ColorID)
	if err != nil {
		log.Fatal("Could not parse GravityID")
	}
	colorChar = ble.NewCharacteristic(colorUUID)
}

func NewTiltClient(address ble.Addr, timeout time.Duration) (*TiltClient, error) {
	tiltClient := &TiltClient{Address: address}
	var profile *ble.Profile

	log.Debugf("[tilt] Connecting to tilt: %s ...", address)
	client, err := ble.Dial(ble.WithSigHandler(context.WithTimeout(context.Background(), timeout)), address)
	if err != nil {
		return tiltClient, err
	}
	defer client.CancelConnection()

	log.Debug("[tilt] Discovering tilt profile...")
	result, err := tiltClient.wrapTimeout(timeout, func() (interface{}, error) {
		return client.DiscoverProfile(false)
	})
	if err != nil {
		return tiltClient, fmt.Errorf("Error discovering tilt profile: %s", err)
	}
	profile = result.(*ble.Profile)

	log.Debug("[tilt] Reading tilt color...")
	result, err = tiltClient.wrapTimeout(timeout, func() (interface{}, error) {
		return tiltClient.GetColor(client, profile)
	})
	if err != nil {
		return tiltClient, fmt.Errorf("[tilt] Error reading tilt color: %s", err)
	}
	tiltClient.Color = result.(string)

	return tiltClient, err
}

func (t *TiltClient) RefreshMetrics(timeout time.Duration) (Metric, error) {
	var profile *ble.Profile
	metric := &Metric{DeviceID: t.Address.String()}

	log.Debug("[tilt] Establishing connection...")
	client, err := ble.Dial(ble.WithSigHandler(context.WithTimeout(context.Background(), timeout)), t.Address)
	if err != nil {
		return *metric, err
	}
	defer client.CancelConnection()

	log.Debug("[tilt] Discovering tilt profile...")
	result, err := t.wrapTimeout(timeout, func() (interface{}, error) {
		return client.DiscoverProfile(false)
	})
	if err != nil {
		return *metric, fmt.Errorf("Error discovering tilt profile: %s", err)
	}
	profile = result.(*ble.Profile)

	log.Debug("[tilt] Reading tilt power...")
	result, err = t.wrapTimeout(timeout, func() (interface{}, error) {
		res := client.ReadRSSI()
		return res, nil
	})
	if err != nil {
		return *metric, fmt.Errorf("Error reading signal power: %s", err)
	}
	metric.Power = result.(int)

	log.Debug("[tilt] Reading tilt battery...")
	result, err = t.wrapTimeout(timeout, func() (interface{}, error) {
		return t.GetBattery(client, profile)
	})
	if err != nil {
		return *metric, fmt.Errorf("Error reading battery: %s", err)
	}
	metric.Battery = result.(int)

	log.Debug("[tilt] Reading tilt temperature...")
	result, err = t.wrapTimeout(timeout, func() (interface{}, error) {
		return t.GetTemperature(client, profile)
	})
	if err != nil {
		return *metric, fmt.Errorf("Error reading temperature: %s", err)
	}
	metric.Temperature = result.(int)

	log.Debug("[tilt] Reading tilt gravity...")
	result, err = t.wrapTimeout(timeout, func() (interface{}, error) {
		return t.GetGravity(client, profile)
	})
	if err != nil {
		return *metric, fmt.Errorf("Error reading gravity: %s", err)
	}
	metric.Gravity = result.(float64)

	return *metric, nil
}

func (t *TiltClient) GetBattery(client ble.Client, profile *ble.Profile) (int, error) {
	if c := profile.Find(batteryChar); c != nil {
		data, err := client.ReadCharacteristic(c.(*ble.Characteristic))
		if err != nil {
			return 0, err
		}
		val, _ := binary.Uvarint(data)
		return strconv.Atoi(fmt.Sprintf("%d", val))
	}
	return 0, fmt.Errorf("Could not find battery characteristic")
}

func (t *TiltClient) GetTemperature(client ble.Client, profile *ble.Profile) (int, error) {
	if c := profile.Find(temperatureChar); c != nil {
		data, err := client.ReadCharacteristic(c.(*ble.Characteristic))
		if err != nil {
			return 0, err
		}
		val, _ := binary.Uvarint(data)
		return strconv.Atoi(fmt.Sprintf("%d", val))
	}
	return 0, fmt.Errorf("Could not find temperature characteristic")
}

func (t *TiltClient) GetGravity(client ble.Client, profile *ble.Profile) (float64, error) {
	if c := profile.Find(gravityChar); c != nil {
		data, err := client.ReadCharacteristic(c.(*ble.Characteristic))
		if err != nil {
			return 0, err
		}
		val, err := strconv.Atoi(fmt.Sprintf("%d", binary.LittleEndian.Uint32(data)))
		return float64(val) / 1000.0, err
	}
	return 0, fmt.Errorf("Could not find gravity characteristic")
}

func (t *TiltClient) GetColor(client ble.Client, profile *ble.Profile) (string, error) {
	if c := profile.Find(colorChar); c != nil {
		data, err := client.ReadCharacteristic(c.(*ble.Characteristic))
		if err != nil {
			return "", err
		}
		switch hex.EncodeToString(data) {
		case ColorRed:
			return "red", nil
		case ColorGreen:
			return "green", nil
		case ColorBlack:
			return "black", nil
		case ColorPurple:
			return "purple", nil
		case ColorOrange:
			return "orange", nil
		case ColorBlue:
			return "blue", nil
		case ColorYellow:
			return "yellow", nil
		case ColorPink:
			return "pink", nil
		}
	}
	return "", fmt.Errorf("Could not determine color")
}

func (t *TiltClient) wrapTimeout(timeout time.Duration, f func() (interface{}, error)) (interface{}, error) {
	var result interface{}
	var err error
	resultChan := make(chan bool, 1)
	defer close(resultChan)

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	go func() {
		result, err = f()
		resultChan<-true
	}()
	select {
	case <-resultChan:
		return result, err
	case <-timer.C:
		fmt.Errorf("timeout")
	}
	return result, nil
}
