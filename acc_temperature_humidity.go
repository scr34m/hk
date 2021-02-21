package main

import (
	"strconv"
	"time"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
)

type TemperatureHumidity struct {
	Accessory         *accessory.Accessory
	TemperatureSensor *service.TemperatureSensor
	HumiditySensor    *service.HumiditySensor
}

func NewAccessoryTemperatureHumidity(conf *ConfigurationDevice) *TemperatureHumidity {
	acc := TemperatureHumidity{}
	acc.Accessory = accessory.New(accessory.Info{Name: conf.Name, SerialNumber: conf.Serialnumber, Manufacturer: conf.Manufacturer, FirmwareRevision: conf.Version}, 24)

	acc.TemperatureSensor = service.NewTemperatureSensor()
	acc.TemperatureSensor.CurrentTemperature.SetMinValue(-100)
	acc.Accessory.AddService(acc.TemperatureSensor.Service)

	acc.HumiditySensor = service.NewHumiditySensor()
	acc.Accessory.AddService(acc.HumiditySensor.Service)

	ticker := time.NewTicker(time.Second * 60)
	go func() {
		for ; true; <-ticker.C {
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
