package main

import (
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/log"
	"github.com/brutella/hc/service"
	"github.com/scr34m/tuya"
)

type TuyaHumidifier_ServiceFan struct {
	Service       *service.Service
	On            *characteristic.On
	RotationSpeed *characteristic.RotationSpeed
}

func NewTuyaHumidifierServiceFan() *TuyaHumidifier_ServiceFan {
	svc := TuyaHumidifier_ServiceFan{}
	svc.Service = service.New(service.TypeFan)

	svc.On = characteristic.NewOn()
	svc.Service.AddCharacteristic(svc.On.Characteristic)

	svc.RotationSpeed = characteristic.NewRotationSpeed()
	svc.RotationSpeed.SetStepValue(50)
	svc.Service.AddCharacteristic(svc.RotationSpeed.Characteristic)
	return &svc
}

type TuyaHumidifier_ServiceLightbulb struct {
	Service    *service.Service
	On         *characteristic.On
	Brightness *characteristic.Brightness
}

func NewTuyaHumidifier_ServiceLightbulb() *TuyaHumidifier_ServiceLightbulb {
	svc := TuyaHumidifier_ServiceLightbulb{}
	svc.Service = service.New(service.TypeLightbulb)

	svc.On = characteristic.NewOn()
	svc.Service.AddCharacteristic(svc.On.Characteristic)

	svc.Brightness = characteristic.NewBrightness()
	svc.Brightness.SetMaxValue(255)
	svc.Service.AddCharacteristic(svc.Brightness.Characteristic)
	return &svc
}

type TuyaHumidifier struct {
	Accessory                       *accessory.Accessory
	TuyaHumidifier_ServiceLightbulb *TuyaHumidifier_ServiceLightbulb
	TuyaHumidifier_ServiceFan       *TuyaHumidifier_ServiceFan
}

func NewAccessoryHumidifier(dm *tuya.DeviceManager, internalname string, conf *ConfigurationDevice) *TuyaHumidifier {
	s := new(ITuyaDeviceHumifier)

	dm.DefineDevice(internalname, conf.Serialnumber, conf.Key, conf.Ip, conf.Version, s)

	d, _ := dm.GetDevice(internalname)
	sw1 := d.(TuyaDeviceHumifier)

	acc := TuyaHumidifier{}
	acc.Accessory = accessory.New(accessory.Info{Name: conf.Name, SerialNumber: conf.Serialnumber, Manufacturer: conf.Manufacturer, FirmwareRevision: conf.Version}, accessory.TypeHumidifier)

	acc.TuyaHumidifier_ServiceLightbulb = NewTuyaHumidifier_ServiceLightbulb()
	acc.Accessory.AddService(acc.TuyaHumidifier_ServiceLightbulb.Service)

	acc.TuyaHumidifier_ServiceFan = NewTuyaHumidifierServiceFan()
	acc.Accessory.AddService(acc.TuyaHumidifier_ServiceFan.Service)

	syncChannel := tuya.MakeSyncChannel()
	d.Subscribe(syncChannel)

	go func() {
		for {
			select {
			case <-syncChannel:
				acc.TuyaHumidifier_ServiceLightbulb.On.SetValue(sw1.Status("11").(bool))
				acc.TuyaHumidifier_ServiceLightbulb.Brightness.SetValue(mapInt(int(sw1.Status("111").(float64)), 0, 255, 0, 100))

				v := sw1.Status("103").(string)
				if v == "off" {
					acc.TuyaHumidifier_ServiceFan.On.SetValue(false)
					acc.TuyaHumidifier_ServiceFan.RotationSpeed.SetValue(0)
				} else if v == "small" {
					acc.TuyaHumidifier_ServiceFan.On.SetValue(true)
					acc.TuyaHumidifier_ServiceFan.RotationSpeed.SetValue(50)
				} else if v == "big" {
					acc.TuyaHumidifier_ServiceFan.On.SetValue(true)
					acc.TuyaHumidifier_ServiceFan.RotationSpeed.SetValue(100)
				}
			}
		}
	}()

	sw1Pending := false

	acc.TuyaHumidifier_ServiceLightbulb.On.OnValueRemoteUpdate(func(on bool) {
		log.Info.Printf("Lightbulb On %v\n", on)

		if sw1Pending {
			log.Info.Printf("Lightbulb working...\n")
			return
		} else {
			sw1Pending = true
		}

		var err error
		if on {
			_, err = sw1.SetW("11", true, 2)
			if err == nil {
				_, err = sw1.SetW("110", "3", 2)
				if err == nil {
					acc.TuyaHumidifier_ServiceLightbulb.On.SetValue(true)
				}
			}
		} else {
			_, err = sw1.SetW("11", false, 2)
			if err == nil {
				acc.TuyaHumidifier_ServiceLightbulb.On.SetValue(false)
			}
		}
		sw1Pending = false

	})
	acc.TuyaHumidifier_ServiceLightbulb.Brightness.OnValueRemoteUpdate(func(value int) {
		log.Info.Printf("Lightbulb Brightness %v\n", value)

		if sw1Pending {
			log.Info.Printf("Lightbulb working...\n")
			return
		} else {
			sw1Pending = true
		}

		if value > 0 {
			var err error
			_, err = sw1.SetW("111", mapInt(value, 0, 100, 0, 255), 2)
			if err == nil {
				acc.TuyaHumidifier_ServiceLightbulb.Brightness.SetValue(value)
			}
		}
		sw1Pending = false
	})

	acc.TuyaHumidifier_ServiceFan.On.OnValueRemoteUpdate(func(on bool) {
		log.Info.Printf("Fan On %v\n", on)

		if sw1Pending {
			log.Info.Printf("Fan working...\n")
			return
		} else {
			sw1Pending = true
		}

		var err error
		if on {
			_, err = sw1.SetW("103", "small", 2)
			if err == nil {
				acc.TuyaHumidifier_ServiceFan.On.SetValue(true)
			}
		} else {
			_, err = sw1.SetW("103", "off", 2)
			if err == nil {
				acc.TuyaHumidifier_ServiceFan.On.SetValue(false)
			}
		}
		sw1Pending = false
	})
	acc.TuyaHumidifier_ServiceFan.RotationSpeed.OnValueRemoteUpdate(func(value float64) {
		log.Info.Printf("Fan RotationSpeed %v\n", value)

		if sw1Pending {
			log.Info.Printf("Fan working...\n")
			return
		} else {
			sw1Pending = true
		}

		var err error
		if value == 100 {
			_, err = sw1.SetW("103", "big", 2)
			if err == nil {
				acc.TuyaHumidifier_ServiceFan.On.SetValue(true)
			}
		} else if value == 50 {
			_, err = sw1.SetW("103", "small", 2)
			if err == nil {
				acc.TuyaHumidifier_ServiceFan.On.SetValue(true)
			}
		}

		sw1Pending = false
	})
	return &acc
}
