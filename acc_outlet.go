package main

import (
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/log"
	"github.com/brutella/hc/service"
	"github.com/scr34m/tuya"
)

type TuyaOutlet_ServiceOutlet struct {
	Service *service.Service
	On      *characteristic.On
}

func NewTuyaOutlet_ServiceOutlet() *TuyaOutlet_ServiceOutlet {
	svc := TuyaOutlet_ServiceOutlet{}
	svc.Service = service.New(service.TypeOutlet)

	svc.On = characteristic.NewOn()
	svc.Service.AddCharacteristic(svc.On.Characteristic)

	return &svc
}

type TuyaOutlet struct {
	Accessory                *accessory.Accessory
	TuyaOutlet_ServiceOutlet *TuyaOutlet_ServiceOutlet
}

func NewAccessoryOutlet(dm *tuya.DeviceManager, internalname string, conf *ConfigurationDevice) *TuyaOutlet {
	s := new(ITuyaDeviceOutlet)

	dm.DefineDevice(internalname, conf.Serialnumber, conf.Key, conf.Ip, conf.Version, s)

	d, _ := dm.GetDevice(internalname)
	sw1 := d.(TuyaDeviceOutlet)

	acc := TuyaOutlet{}
	acc.Accessory = accessory.New(accessory.Info{Name: conf.Name, SerialNumber: conf.Serialnumber, Manufacturer: conf.Manufacturer, FirmwareRevision: conf.Version}, accessory.TypeOutlet)

	acc.TuyaOutlet_ServiceOutlet = NewTuyaOutlet_ServiceOutlet()
	acc.Accessory.AddService(acc.TuyaOutlet_ServiceOutlet.Service)

	syncChannel := tuya.MakeSyncChannel()
	d.Subscribe(syncChannel)

	go func() {
		for {
			select {
			case <-syncChannel:
				acc.TuyaOutlet_ServiceOutlet.On.SetValue(sw1.Status("1").(bool))
			}
		}
	}()

	sw1Pending := false

	acc.TuyaOutlet_ServiceOutlet.On.OnValueRemoteUpdate(func(on bool) {
		log.Info.Printf("Outletbulb On %v\n", on)

		if sw1Pending {
			log.Info.Printf("Outletbulb working...\n")
			return
		}

		// bulb status is same then stop
		if on == sw1.Status("1").(bool) {
			return
		}

		sw1Pending = true

		var err error
		if on {
			_, err = sw1.SetW("1", true, 2)
			if err == nil {
				acc.TuyaOutlet_ServiceOutlet.On.SetValue(true)
			}
		} else {
			_, err = sw1.SetW("1", false, 2)
			if err == nil {
				acc.TuyaOutlet_ServiceOutlet.On.SetValue(false)
			}
		}
		sw1Pending = false

	})
	return &acc
}
