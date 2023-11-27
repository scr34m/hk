package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/log"
	"github.com/brutella/hc/service"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type HeatingSystem struct {
	Accessory                 *accessory.Accessory
	TemperatureForwardSensor  *service.TemperatureSensor
	TemperatureBackwardSensor *service.TemperatureSensor

	internalname string
	mqtt_cli     mqtt.Client
	url          string
}

func (s *HeatingSystem) pub(name string, payload interface{}) {
	token := s.mqtt_cli.Publish("hk/"+s.internalname+"/"+name, 0, true, fmt.Sprintf("%v", payload))
	token.Wait()
	if token.Error() != nil {
		log.Info.Println(token.Error())
	}
}

func (s *HeatingSystem) init(internalname string, mqtt_cli mqtt.Client, url string) {
	s.internalname = internalname
	s.mqtt_cli = mqtt_cli
	s.url = url

	ticker := time.NewTicker(time.Second * 60)
	go func() {
		for ; true; <-ticker.C {
			var result map[string]interface{}
			err := getJSON(s.url, &result)
			if err == nil {
				if val, ok := result["forward"]; ok {
					f, _ := strconv.ParseFloat(val.(string), 64)
					s.TemperatureForwardSensor.CurrentTemperature.SetValue(f)
					s.pub("forward", f)
				}
				if val, ok := result["backward"]; ok {
					f, _ := strconv.ParseFloat(val.(string), 64)
					s.TemperatureBackwardSensor.CurrentTemperature.SetValue(f)
					s.pub("backward", f)
				}
			}
		}
	}()
}

func NewAccessoryHeatingSystem(conf *ConfigurationDevice, mqtt_cli mqtt.Client) *HeatingSystem {
	acc := HeatingSystem{}
	acc.Accessory = accessory.New(accessory.Info{Name: conf.Name, SerialNumber: conf.Serialnumber, Manufacturer: conf.Manufacturer, FirmwareRevision: conf.Version}, 22)

	acc.TemperatureForwardSensor = service.NewTemperatureSensor()
	acc.TemperatureForwardSensor.CurrentTemperature.SetMinValue(0)
	acc.Accessory.AddService(acc.TemperatureForwardSensor.Service)

	acc.TemperatureBackwardSensor = service.NewTemperatureSensor()
	acc.TemperatureBackwardSensor.CurrentTemperature.SetMinValue(0)
	acc.Accessory.AddService(acc.TemperatureBackwardSensor.Service)

	acc.init("heating_system", mqtt_cli, conf.Url)

	return &acc
}
