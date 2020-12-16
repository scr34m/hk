package main

import (
	"strconv"
	"time"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
)

type SensorTempHum struct {
	Accessory         *accessory.Accessory
	TemperatureSensor *service.TemperatureSensor
	HumiditySensor    *service.HumiditySensor
}

func NewAccessorySensorTempHum(conf *ConfigurationDevice) *SensorTempHum {
	acc := SensorTempHum{}
	acc.Accessory = accessory.New(accessory.Info{Name: conf.Name, SerialNumber: conf.Serialnumber, Manufacturer: "Sencor", FirmwareRevision: conf.Version}, accessory.TypeHumidifier)

	acc.TemperatureSensor = service.NewTemperatureSensor()
	acc.Accessory.AddService(acc.TemperatureSensor.Service)

	acc.HumiditySensor = service.NewHumiditySensor()
	acc.Accessory.AddService(acc.HumiditySensor.Service)

	ticker := time.NewTicker(time.Second * 2)
	go func() {
		for _ = range ticker.C {
			var result map[string]interface{}
			err := getJSON(conf.Url, &result)
			if err == nil {
				if val, ok := result["temperature"]; ok {
					f, _ := strconv.ParseFloat(val.(string), 64)
					acc.TemperatureSensor.CurrentTemperature.SetValue(f)
				}
				if val, ok := result["humidity"]; ok {
					f, _ := strconv.ParseFloat(val.(string), 64)
					acc.HumiditySensor.CurrentRelativeHumidity.SetValue(f)
				}
			}
		}
	}()

	return &acc
}
