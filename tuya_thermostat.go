package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/scr34m/tuya"
)

type TuyaDeviceThermostat interface {
	Set(string, interface{}) error
	SetW(string, interface{}, time.Duration) (interface{}, error)
	Status(string) interface{}
}

type ITuyaDeviceThermostat struct {
	tuya.BaseDevice
	// 1:true on / off
	// 101:AAYoAAgeHgseHg0eABEsABYeAAYoAAgoHgsoHg0oABEoABYeAAYoAAgoHgsoHg0oABEoABYe ?
	// 102:0 on / off relay (43,44,45,46)
	// 103:0 ?
	// 104:true ?
	// 2:50 Heating temerature (celsius / 2)
	// 3:51 Floor temerature (celsius / 2)
	// 4:1 on/off progam mode
	// 5:false on/off eco (2:40)
	// 6:false on/off close
	status map[string]interface{}
}

func (s *ITuyaDeviceThermostat) Set(dps string, value interface{}) error {
	m := s.App.MakeBaseMsg()
	m["dps"] = map[string]interface{}{dps: value}
	return s.App.SendEncryptedCommand(tuya.CodeMsgSet, m)
}

func (s *ITuyaDeviceThermostat) SetW(dps string, value interface{}, delay time.Duration) (interface{}, error) {
	c := tuya.MakeSyncChannel()
	k := s.Subscribe(c)
	defer s.Unsubscribe(k)

	deadLine := time.Now().Add(delay * time.Second)
	m := s.App.MakeBaseMsg()
	m["dps"] = map[string]interface{}{dps: value}
	err := s.App.SendEncryptedCommand(tuya.CodeMsgSet, m)
	if err != nil {
		return nil, err
	}
	for {
		select {
		case synMsg := <-c:
			if synMsg.Code == tuya.CodeMsgStatus || synMsg.Code == tuya.CodeMsgAutoStatus {
				s := s.Status(dps)
				return s, err
			}
		case <-time.After(deadLine.Sub(time.Now())):
			return nil, errors.New("Timeout")
		}
	}
}

func (s *ITuyaDeviceThermostat) Status(key string) interface{} {
	if val, ok := s.status[key]; ok {
		return val
	} else {
		return nil
	}
}

func (s *ITuyaDeviceThermostat) ProcessResponse(code int, data []byte) {
	switch {
	case len(data) == 0:
		return
	case code == 7:
		return
	case code == 9:
		return
	}

	var r map[string]interface{}
	err := json.Unmarshal(data, &r)
	if err != nil {
		log.Println("JSON decode error")
		return
	}

	if val, ok := r["dps"]; ok {
		v2 := val.(map[string]interface{})
		for k, v := range v2 {
			s.status[k] = v
		}
	}
	log.Printf("TuyaDeviceThermostat: %v\n", r)
	s.Notify(code, s)
}

func (s *ITuyaDeviceThermostat) Configure(a *tuya.Appliance, c *tuya.ConfigurationData) {
	s.status = make(map[string]interface{})
	s.Init("TuyaDeviceThermostat", a, c)
}
