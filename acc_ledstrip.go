package main

import (
	"fmt"
	"strconv"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/log"
	"github.com/brutella/hc/service"
	"github.com/scr34m/tuya"
)

type TuyaLedstrip_Service struct {
	Service    *service.Service
	On         *characteristic.On
	Brightness *characteristic.Brightness
	Saturation *characteristic.Saturation
	Hue        *characteristic.Hue
}

func NewTuyaLedstrip_Service() *TuyaLedstrip_Service {
	svc := TuyaLedstrip_Service{}
	svc.Service = service.New(service.TypeLightbulb)

	svc.On = characteristic.NewOn()
	svc.Service.AddCharacteristic(svc.On.Characteristic)

	svc.Hue = characteristic.NewHue()
	svc.Hue.SetValue(0)
	svc.Hue.SetMaxValue(360)
	svc.Service.AddCharacteristic(svc.Hue.Characteristic)

	svc.Saturation = characteristic.NewSaturation()
	svc.Saturation.SetValue(0)
	svc.Saturation.SetMaxValue(100)
	svc.Service.AddCharacteristic(svc.Saturation.Characteristic)

	svc.Brightness = characteristic.NewBrightness()
	svc.Brightness.SetValue(0)
	svc.Brightness.SetMaxValue(255)
	svc.Service.AddCharacteristic(svc.Brightness.Characteristic)

	return &svc
}

type TuyaLedstrip struct {
	Accessory            *accessory.Accessory
	TuyaLedstrip_Service *TuyaLedstrip_Service
}

func convert_hsb(hue float64, saturation float64, brightness float64) string {
	var v int64

	v = int64(hue)
	v = v << 16
	v = v + int64(mapInt(int(saturation), 0, 100, 0, 1000))
	v = v << 16
	v = v + int64(mapInt(int(brightness), 0, 100, 0, 1000))

	return fmt.Sprintf("%012s", strconv.FormatInt(v, 16))
}

func NewAccessoryLedstrip(dm *tuya.DeviceManager, internalname string, conf *ConfigurationDevice) *TuyaLedstrip {
	s := new(ITuyaDeviceLedstrip)

	dm.DefineDevice(internalname, conf.Serialnumber, conf.Key, conf.Ip, conf.Version, s)

	d, _ := dm.GetDevice(internalname)
	sw1 := d.(TuyaDeviceLedstrip)

	acc := TuyaLedstrip{}
	acc.Accessory = accessory.New(accessory.Info{Name: conf.Name, SerialNumber: conf.Serialnumber, Manufacturer: conf.Manufacturer, FirmwareRevision: conf.Version}, accessory.TypeLightbulb)

	acc.TuyaLedstrip_Service = NewTuyaLedstrip_Service()
	acc.Accessory.AddService(acc.TuyaLedstrip_Service.Service)

	syncChannel := tuya.MakeSyncChannel()
	d.Subscribe(syncChannel)

	var hue float64
	var saturation float64
	var brightness float64

	sw1Pending := true

	go func() {
		for {
			select {
			case <-syncChannel:
				acc.TuyaLedstrip_Service.On.SetValue(sw1.Status("20").(bool))

				vs := sw1.Status("24").(string)
				log.Info.Printf("%v\n", vs)

				v, _ := strconv.ParseInt(vs, 16, 64)

				hue = float64(int((v >> 32) & 0xffff))
				saturation = float64(mapInt(int(v>>16)&0xffff, 0, 1000, 0, 100))
				brightness = float64(mapInt(int(v&0xffff), 0, 1000, 0, 100))

				log.Info.Printf("hue %v, saturation %v, brightness %v\n", hue, saturation, brightness)

				acc.TuyaLedstrip_Service.Hue.SetValue(hue)
				acc.TuyaLedstrip_Service.Saturation.SetValue(saturation)
				acc.TuyaLedstrip_Service.Brightness.SetValue(int(brightness))

				sw1Pending = false
			}
		}
	}()

	acc.TuyaLedstrip_Service.On.OnValueRemoteUpdate(func(on bool) {
		log.Info.Printf("Ledstrip On %v\n", on)

		if sw1Pending {
			log.Info.Printf("Ledstrip working...\n")
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
				acc.TuyaLedstrip_Service.On.SetValue(true)
			}
		} else {
			_, err = sw1.SetW("20", false, 2)
			if err == nil {
				acc.TuyaLedstrip_Service.On.SetValue(false)
			}
		}
		sw1Pending = false
	})

	acc.TuyaLedstrip_Service.Brightness.OnValueRemoteUpdate(func(value int) {
		log.Info.Printf("Ledstrip Brightness %v\n", value)

		if sw1Pending {
			log.Info.Printf("Ledstrip working...\n")
			return
		} else {
			sw1Pending = true
		}

		var err error
		brightness = float64(value)
		_, err = sw1.SetW("24", convert_hsb(hue, saturation, brightness), 2)
		if err == nil {
			acc.TuyaLedstrip_Service.Brightness.SetValue(value)
		}
		sw1Pending = false
	})

	acc.TuyaLedstrip_Service.Hue.OnValueRemoteUpdate(func(value float64) {
		log.Info.Printf("Ledstrip Hue %v\n", value)

		if sw1Pending {
			log.Info.Printf("Ledstrip working...\n")
			return
		} else {
			sw1Pending = true
		}

		var err error
		hue = value
		_, err = sw1.SetW("24", convert_hsb(hue, saturation, brightness), 2)
		if err == nil {
			acc.TuyaLedstrip_Service.Hue.SetValue(value)
		}
		sw1Pending = false
	})

	acc.TuyaLedstrip_Service.Saturation.OnValueRemoteUpdate(func(value float64) {
		log.Info.Printf("Ledstrip Saturation %v\n", value)

		if sw1Pending {
			log.Info.Printf("Ledstrip working...\n")
			return
		} else {
			sw1Pending = true
		}

		var err error
		saturation = value
		_, err = sw1.SetW("24", convert_hsb(hue, saturation, brightness), 2)
		if err == nil {
			acc.TuyaLedstrip_Service.Saturation.SetValue(value)
		}
		sw1Pending = false
	})

	return &acc
}
