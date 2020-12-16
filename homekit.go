package main

import (
	"context"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/log"
)

type homekit struct {
	Pin       string
	Type      string
	Device    ConfigurationDevice
	transport hc.Transport
}

func (h *homekit) Init() error {
	log.Info.Printf("Homekit accessory initialisation for %s\n", h.Type)

	var a *accessory.Accessory

	switch h.Type {
	case "humifier":
		hum1 := NewAccessoryHumidifier(&h.Device)
		a = hum1.Accessory
	case "temperature_humidity":
		hum1 := NewAccessorySensorTempHum(&h.Device)
		a = hum1.Accessory
	}

	// TODO use StoragePath
	t, err := hc.NewIPTransport(hc.Config{Pin: h.Pin}, a)
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
	return &homekit{Pin: a.Pin, Type: a.Type, Device: a.Device}
}
