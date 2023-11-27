package main

import (
	"fmt"
	"strconv"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/log"
	"github.com/brutella/hc/service"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/scr34m/tuya"
)

type TuyaThermostat_ServiceThermostat struct {
	Service *service.Service

	On                         *characteristic.On
	CurrentHeatingCoolingState *characteristic.CurrentHeatingCoolingState
	TargetHeatingCoolingState  *characteristic.TargetHeatingCoolingState
	CurrentTemperature         *characteristic.CurrentTemperature
	TargetTemperature          *characteristic.TargetTemperature
	TemperatureDisplayUnits    *characteristic.TemperatureDisplayUnits
}

func NewTuyaThermostat_ServiceThermostat() *TuyaThermostat_ServiceThermostat {
	svc := TuyaThermostat_ServiceThermostat{}
	svc.Service = service.New(service.TypeThermostat)

	svc.On = characteristic.NewOn()
	svc.Service.AddCharacteristic(svc.On.Characteristic)

	svc.CurrentHeatingCoolingState = characteristic.NewCurrentHeatingCoolingState()
	svc.Service.AddCharacteristic(svc.CurrentHeatingCoolingState.Characteristic)

	svc.TargetHeatingCoolingState = characteristic.NewTargetHeatingCoolingState()
	svc.Service.AddCharacteristic(svc.TargetHeatingCoolingState.Characteristic)

	svc.CurrentTemperature = characteristic.NewCurrentTemperature()
	svc.CurrentTemperature.SetValue(20)
	svc.Service.AddCharacteristic(svc.CurrentTemperature.Characteristic)

	svc.TargetTemperature = characteristic.NewTargetTemperature()
	svc.TargetTemperature.SetStepValue(0.5)
	svc.Service.AddCharacteristic(svc.TargetTemperature.Characteristic)

	svc.TemperatureDisplayUnits = characteristic.NewTemperatureDisplayUnits()
	svc.Service.AddCharacteristic(svc.TemperatureDisplayUnits.Characteristic)

	return &svc
}

type TuyaThermostat struct {
	Accessory                        *accessory.Accessory
	TuyaThermostat_ServiceThermostat *TuyaThermostat_ServiceThermostat

	device       TuyaDeviceThermostat
	pending      bool
	internalname string
	mqtt_cli     mqtt.Client
}

func (s *TuyaThermostat) pub(name string, payload interface{}) {
	token := s.mqtt_cli.Publish("hk/"+s.internalname+"/"+name, 0, true, fmt.Sprintf("%v", payload))
	token.Wait()
	if token.Error() != nil {
		log.Info.Println(token.Error())
	}
}

func (s *TuyaThermostat) OnUpdate(on bool) {
	log.Info.Printf("Terhmostat on")

	s.pub("on", on)

	if s.pending {
		log.Info.Printf("Terhmostat working...\n")
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
			s.TuyaThermostat_ServiceThermostat.On.SetValue(true)
		}
	} else {
		_, err = s.device.SetW("1", false, 2)
		if err == nil {
			s.TuyaThermostat_ServiceThermostat.On.SetValue(false)
		}
	}
	s.pending = false
}

func (s *TuyaThermostat) TargetTemperatureSub(client mqtt.Client, msg mqtt.Message) {
	value, _ := strconv.ParseFloat(string(msg.Payload()), 64)
	log.Info.Printf("TargetTemperatureSub %v\n", value)
	s.TargetTemperatureSet(value)
}

func (s *TuyaThermostat) TargetTemperatureUpdate(value float64) {
	log.Info.Printf("Terhmostat TargetTemperature %v\n", value)
	s.pub("heating/set", value)
	s.TargetTemperatureSet(value)
}

func (s *TuyaThermostat) TargetTemperatureSet(value float64) {
	if s.pending {
		log.Info.Printf("Terhmostat working...\n")
		return
	}

	s.pending = true
	var err error
	_, err = s.device.SetW("2", value*2, 2)
	if err == nil {
		s.TuyaThermostat_ServiceThermostat.TargetTemperature.SetValue(value)
	}
	s.pending = false
}

func (s *TuyaThermostat) CurrentTemperatureUpdate(value float64) {
	log.Info.Printf("Terhmostat CurrentTemperature %v\n", value)

	s.pub("temperature", value)
}

func (s *TuyaThermostat) init(internalname string, mqtt_cli mqtt.Client) {
	s.internalname = internalname
	s.mqtt_cli = mqtt_cli

	d, _ := dm.GetDevice(s.internalname)
	s.device = d.(TuyaDeviceThermostat)

	s.pending = false

	syncChannel := tuya.MakeSyncChannel()
	d.Subscribe(syncChannel)

	go func() {
		for {
			select {
			case <-syncChannel:
				s.TuyaThermostat_ServiceThermostat.On.SetValue(s.device.Status("1").(bool))
				s.TuyaThermostat_ServiceThermostat.TargetTemperature.SetValue(s.device.Status("2").(float64) / 2)
				s.TuyaThermostat_ServiceThermostat.CurrentTemperature.SetValue(s.device.Status("3").(float64) / 2)

				s.pub("on", s.device.Status("1").(bool))
				s.pub("heating/get", s.device.Status("2").(float64)/2)
				s.pub("temperature", s.device.Status("3").(float64)/2)
			}
		}
	}()

	topic := "hk/" + s.internalname + "/heating/set"
	token := s.mqtt_cli.Subscribe(topic, 1, s.TargetTemperatureSub)
	token.Wait()

	log.Info.Printf("Subscribed to topic: %s\n", topic)

	s.TuyaThermostat_ServiceThermostat.On.OnValueRemoteUpdate(s.OnUpdate)
	s.TuyaThermostat_ServiceThermostat.TargetTemperature.OnValueRemoteUpdate(s.TargetTemperatureUpdate)
	s.TuyaThermostat_ServiceThermostat.CurrentTemperature.OnValueRemoteUpdate(s.CurrentTemperatureUpdate)
}

func NewAccessoryThermostat(dm *tuya.DeviceManager, internalname string, conf *ConfigurationDevice, mqtt_cli mqtt.Client) *TuyaThermostat {
	s := new(ITuyaDeviceThermostat)

	dm.DefineDevice(internalname, conf.Serialnumber, conf.Key, conf.Ip, conf.Version, s)

	acc := TuyaThermostat{}
	acc.Accessory = accessory.New(accessory.Info{Name: conf.Name, SerialNumber: conf.Serialnumber, Manufacturer: conf.Manufacturer, FirmwareRevision: conf.Version}, accessory.TypeThermostat)
	acc.TuyaThermostat_ServiceThermostat = NewTuyaThermostat_ServiceThermostat()
	acc.Accessory.AddService(acc.TuyaThermostat_ServiceThermostat.Service)

	acc.init(internalname, mqtt_cli)

	return &acc
}
