package main

import (
	"strconv"

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

func (s *HeatingSystem) OnForwardSet(client mqtt.Client, msg mqtt.Message) {
	value, _ := strconv.ParseFloat(string(msg.Payload()), 64)
	log.Info.Printf("OnForwardSet %v\n", value)
	s.TemperatureForwardSensor.CurrentTemperature.SetValue(value)
}

func (s *HeatingSystem) OnBackwardSet(client mqtt.Client, msg mqtt.Message) {
	value, _ := strconv.ParseFloat(string(msg.Payload()), 64)
	log.Info.Printf("OnBackwardSet %v\n", value)
	s.TemperatureBackwardSensor.CurrentTemperature.SetValue(value)
}

func (s *HeatingSystem) init(internalname string, mqtt_cli mqtt.Client, url string) {
	s.internalname = internalname
	s.mqtt_cli = mqtt_cli
	s.url = url

	topic := "hk/" + s.internalname + "/forward"
	token := s.mqtt_cli.Subscribe(topic, 1, s.OnForwardSet)
	token.Wait()

	log.Info.Printf("Subscribed to topic: %s\n", topic)

	topic = "hk/" + s.internalname + "/backward"
	token = s.mqtt_cli.Subscribe(topic, 1, s.OnBackwardSet)
	token.Wait()

	log.Info.Printf("Subscribed to topic: %s\n", topic)
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
