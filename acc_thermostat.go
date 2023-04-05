package main

import (
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/log"
	"github.com/brutella/hc/service"
	"github.com/scr34m/tuya"
)

type TuyaThermostat_ServiceThermostat struct {
	Service           *service.Service

	On                *characteristic.On
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
	Accessory                *accessory.Accessory
	TuyaThermostat_ServiceThermostat *TuyaThermostat_ServiceThermostat
}

func NewAccessoryThermostat(dm *tuya.DeviceManager, internalname string, conf *ConfigurationDevice) *TuyaThermostat {
	s := new(ITuyaDeviceThermostat)

	dm.DefineDevice(internalname, conf.Serialnumber, conf.Key, conf.Ip, conf.Version, s)

	d, _ := dm.GetDevice(internalname)
	sw1 := d.(TuyaDeviceThermostat)

	acc := TuyaThermostat{}
	acc.Accessory = accessory.New(accessory.Info{Name: conf.Name, SerialNumber: conf.Serialnumber, Manufacturer: conf.Manufacturer, FirmwareRevision: conf.Version}, accessory.TypeThermostat)

	acc.TuyaThermostat_ServiceThermostat = NewTuyaThermostat_ServiceThermostat()
	acc.Accessory.AddService(acc.TuyaThermostat_ServiceThermostat.Service)

	syncChannel := tuya.MakeSyncChannel()
	d.Subscribe(syncChannel)

	go func() {
		for {
			select {
			case <-syncChannel:
				acc.TuyaThermostat_ServiceThermostat.On.SetValue(sw1.Status("1").(bool))
				acc.TuyaThermostat_ServiceThermostat.CurrentTemperature.SetValue(sw1.Status("3").(float64) / 2)
			}
		}
	}()

	sw1Pending := false

	acc.TuyaThermostat_ServiceThermostat.On.OnValueRemoteUpdate(func(on bool) {
		log.Info.Printf("Terhmostat on")

		if sw1Pending {
			log.Info.Printf("Terhmostat working...\n")
			return
		}

		if on == sw1.Status("1").(bool) {
			return
		}

		sw1Pending = true

		var err error
		if on {
			_, err = sw1.SetW("1", true, 2)
			if err == nil {
				acc.TuyaThermostat_ServiceThermostat.On.SetValue(true)
			}
		} else {
			_, err = sw1.SetW("1", false, 2)
			if err == nil {
				acc.TuyaThermostat_ServiceThermostat.On.SetValue(false)
			}
		}
		sw1Pending = false
	})

	acc.TuyaThermostat_ServiceThermostat.TargetTemperature.OnValueRemoteUpdate(func(value float64) {
		log.Info.Printf("Terhmostat TargetTemperature %v\n", value)

		if sw1Pending {
			log.Info.Printf("Terhmostat working...\n")
			return
		} else {
			sw1Pending = true
		}

		var err error
		_, err = sw1.SetW("2", value * 2, 2)
		if err == nil {
			acc.TuyaThermostat_ServiceThermostat.TargetTemperature.SetValue(value)
		}

		sw1Pending = false
	})

	return &acc
}
