package main

import (
	"context"

    mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/log"
	"github.com/scr34m/tuya"
)

type homekit struct {
	Pin          string
	Type         string
	Internalname string
	Device       ConfigurationDevice
	transport    hc.Transport
}

// TODO this is ugly
var dm *tuya.DeviceManager

func (h *homekit) Init(mqtt_cli mqtt.Client) error {
	log.Info.Printf("Homekit accessory initialisation for %s\n", h.Type)

	var a *accessory.Accessory

	if dm == nil {
		dm = tuya.NewDeviceManagerRaw()
	}

	switch h.Type {
	case "humifier":
		hum1 := NewAccessoryHumidifier(dm, h.Internalname, &h.Device)
		a = hum1.Accessory
	case "light":
		hum1 := NewAccessoryLight(dm, h.Internalname, &h.Device)
		a = hum1.Accessory
	case "ledstrip":
		hum1 := NewAccessoryLedstrip(dm, h.Internalname, &h.Device)
		a = hum1.Accessory
	case "outlet":
		hum1 := NewAccessoryOutlet(dm, h.Internalname, &h.Device, mqtt_cli)
		a = hum1.Accessory
	case "thermostat":
		hum1 := NewAccessoryThermostat(dm, h.Internalname, &h.Device, mqtt_cli)
		a = hum1.Accessory
	case "temperature_humidity":
		hum1 := NewAccessoryTemperatureHumidity(&h.Device)
		a = hum1.Accessory
	case "heating_system":
		hum1 := NewAccessoryHeatingSystem(&h.Device, mqtt_cli)
		a = hum1.Accessory
	default:
		panic(h.Type)
	}

	t, err := hc.NewIPTransport(hc.Config{Pin: h.Pin, StoragePath: h.Internalname}, a)
	if err != nil {
		return err
	}

	h.transport = t

	go h.transport.Start()

	return nil
}

func (h *homekit) Start(ctx context.Context) {
	<-ctx.Done()
	log.Info.Printf("Homekit accessory shutting down for %s\n", h.Type)
	<-h.transport.Stop()
}

func makeHomekit(a ConfigurationAccessory) *homekit {
	return &homekit{Pin: a.Pin, Type: a.Type, Device: a.Device, Internalname: a.Internalname}
}
