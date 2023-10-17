package main

import (
	"time"
	"fmt"

    mqtt "github.com/eclipse/paho.mqtt.golang"
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

	device TuyaDeviceOutlet
	pending bool
	internalname string
	mqtt_cli mqtt.Client
}

func (s *TuyaOutlet) pub(name string, payload interface{}) {
    token := s.mqtt_cli.Publish("hk/" + s.internalname + "/" + name, 0, true, fmt.Sprintf("%v", payload))
    token.Wait()
    if token.Error() != nil {
	    log.Info.Println(token.Error())
    } else {
	    log.Info.Printf("Published to topic: %v payload: %v\n", "hk/" + s.internalname + "/" + name, payload)
    }
}

func (s *TuyaOutlet) OnUpdate(on bool) {
	log.Info.Printf("Outlet on")

	s.pub("on", on)

	if s.pending {
		log.Info.Printf("Outlet working...\n")
		return
	}

	if on == s.device.Status("1").(bool) {
		return
	}

	s.pending = true
	var err error
	if on {
		_, err = s.device.SetW("1", true, 2)
		if err == nil {
			s.TuyaOutlet_ServiceOutlet.On.SetValue(true)
		}
	} else {
		_, err = s.device.SetW("1", false, 2)
		if err == nil {
			s.TuyaOutlet_ServiceOutlet.On.SetValue(false)
		}
	}
	s.pending = false
}

func (s *TuyaOutlet) init(internalname string, mqtt_cli mqtt.Client, conf *ConfigurationDevice) {
	s.internalname = internalname
	s.mqtt_cli = mqtt_cli

	d, _ := dm.GetDevice(internalname)
	s.device = d.(TuyaDeviceOutlet)

	s.pending = false	

	syncChannel := tuya.MakeSyncChannel()
	d.Subscribe(syncChannel)

	go func() {
		for {
			select {
			case <-syncChannel:
				s.TuyaOutlet_ServiceOutlet.On.SetValue(s.device.Status("1").(bool))
			}
		}
	}()


	if conf.Monitor_18 || conf.Monitor_19 {
		ticker := time.NewTicker(time.Second * 20)
		go func() {
			for ; true; <-ticker.C {
				if (conf.Monitor_18) {
					s.pub("current", s.device.Status("18"))
				}
				if (conf.Monitor_19) {
					s.pub("power", s.device.Status("19"))
				}
			}
		}()
	}

	s.TuyaOutlet_ServiceOutlet.On.OnValueRemoteUpdate(s.OnUpdate)
}


func NewAccessoryOutlet(dm *tuya.DeviceManager, internalname string, conf *ConfigurationDevice, mqtt_cli mqtt.Client) *TuyaOutlet {
	s := new(ITuyaDeviceOutlet)

	dm.DefineDevice(internalname, conf.Serialnumber, conf.Key, conf.Ip, conf.Version, s)

	acc := TuyaOutlet{}
	acc.Accessory = accessory.New(accessory.Info{Name: conf.Name, SerialNumber: conf.Serialnumber, Manufacturer: conf.Manufacturer, FirmwareRevision: conf.Version}, accessory.TypeOutlet)
	acc.TuyaOutlet_ServiceOutlet = NewTuyaOutlet_ServiceOutlet()
	acc.Accessory.AddService(acc.TuyaOutlet_ServiceOutlet.Service)

	acc.init(internalname, mqtt_cli, conf)

	return &acc
}



