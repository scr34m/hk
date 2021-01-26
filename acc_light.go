package main

import (
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/log"
	"github.com/brutella/hc/service"
	"github.com/scr34m/tuya"
)

type TuyaLight_ServiceLightbulb struct {
	Service    *service.Service
	On         *characteristic.On
	Brightness *characteristic.Brightness
	Saturation *characteristic.Saturation
	Hue        *characteristic.Hue
}

func NewTuyaLight_ServiceLightbulb() *TuyaLight_ServiceLightbulb {
	svc := TuyaLight_ServiceLightbulb{}
	svc.Service = service.New(service.TypeLightbulb)

	svc.On = characteristic.NewOn()
	svc.Service.AddCharacteristic(svc.On.Characteristic)

	svc.Brightness = characteristic.NewBrightness()
	svc.Brightness.SetValue(0)
	svc.Brightness.SetMaxValue(100)
	svc.Service.AddCharacteristic(svc.Brightness.Characteristic)

	return &svc
}

type TuyaLight struct {
	Accessory                  *accessory.Accessory
	TuyaLight_ServiceLightbulb *TuyaLight_ServiceLightbulb
}

func NewAccessoryLight(dm *tuya.DeviceManager, internalname string, conf *ConfigurationDevice) *TuyaLight {
	s := new(ITuyaDeviceLight)

	dm.DefineDevice(internalname, conf.Serialnumber, conf.Key, conf.Ip, conf.Version, s)

	d, _ := dm.GetDevice(internalname)
	sw1 := d.(TuyaDeviceLight)

	acc := TuyaLight{}
	acc.Accessory = accessory.New(accessory.Info{Name: conf.Name, SerialNumber: conf.Serialnumber, Manufacturer: conf.Manufacturer, FirmwareRevision: conf.Version}, accessory.TypeLightbulb)

	acc.TuyaLight_ServiceLightbulb = NewTuyaLight_ServiceLightbulb()
	acc.Accessory.AddService(acc.TuyaLight_ServiceLightbulb.Service)

	syncChannel := tuya.MakeSyncChannel()
	d.Subscribe(syncChannel)

	go func() {
		for {
			select {
			case <-syncChannel:
				acc.TuyaLight_ServiceLightbulb.On.SetValue(sw1.Status("20").(bool))
				v := mapInt(int(sw1.Status("22").(float64)), 0, 255, 0, 100)
				log.Info.Printf("%v\n", v)
				acc.TuyaLight_ServiceLightbulb.Brightness.SetValue(v)
			}
		}
	}()

	sw1Pending := false

	acc.TuyaLight_ServiceLightbulb.On.OnValueRemoteUpdate(func(on bool) {
		log.Info.Printf("Lightbulb On %v\n", on)

		if sw1Pending {
			log.Info.Printf("Lightbulb working...\n")
			return
		}

		// bulb status is same then stop
		if on == sw1.Status("20").(bool) {
			return
		}

		sw1Pending = true

		var err error
		if on {
			_, err = sw1.SetW("20", true, 2)
			if err == nil {
				acc.TuyaLight_ServiceLightbulb.On.SetValue(true)
			}
		} else {
			_, err = sw1.SetW("20", false, 2)
			if err == nil {
				acc.TuyaLight_ServiceLightbulb.On.SetValue(false)
			}
		}
		sw1Pending = false

	})
	acc.TuyaLight_ServiceLightbulb.Brightness.OnValueRemoteUpdate(func(value int) {
		log.Info.Printf("Lightbulb Brightness %v\n", value)

		if sw1Pending {
			log.Info.Printf("Lightbulb working...\n")
			return
		} else {
			sw1Pending = true
		}

		if value > 0 {
			var err error
			v := mapInt(value, 0, 100, 0, 255)
			log.Info.Printf("%v\n", v)
			_, err = sw1.SetW("22", v, 2)
			if err == nil {
				acc.TuyaLight_ServiceLightbulb.Brightness.SetValue(value)
			}
		}
		sw1Pending = false
	})
	return &acc
}
