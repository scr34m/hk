package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/scr34m/tuya"
)

type TuyaDeviceHumifier interface {
	Set(string, interface{}) error
	SetW(string, interface{}, time.Duration) (interface{}, error)
	Status(string) interface{}
}

type ITuyaDeviceHumifier struct {
	tuya.BaseDevice
	// 1: power
	// 11: light
	// 103: dp_mist_grade: off, small, big
	// 108: colour_data:00ffde00acffff
	// 109: work_mode: white, colour, scene, scene1, scene2, scene3, scene4
	// 110: lightmode: 1 (gradient), 2 (fixed), 3 (night)
	// 111: setlight (brightness)
	// 12: fault
	// 13: countdown: 0, 1, 2, 3
	// 14: countdown_left (0-360)
	status map[string]interface{}
}

func (s *ITuyaDeviceHumifier) Set(dps string, value interface{}) error {
	m := s.App.MakeBaseMsg()
	m["dps"] = map[string]interface{}{dps: value}
	return s.App.SendEncryptedCommand(tuya.CodeMsgSet, m)
}

func (s *ITuyaDeviceHumifier) SetW(dps string, value interface{}, delay time.Duration) (interface{}, error) {
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

func (s *ITuyaDeviceHumifier) Status(key string) interface{} {
	if val, ok := s.status[key]; ok {
		return val
	} else {
		return nil
	}
}

func (s *ITuyaDeviceHumifier) ProcessResponse(code int, data []byte) {
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
	log.Printf("TuyaDeviceHumifier: %v\n", r)
	s.Notify(code, s)
}

func (s *ITuyaDeviceHumifier) Configure(a *tuya.Appliance, c *tuya.ConfigurationData) {
	s.status = make(map[string]interface{})
	s.Init("TuyaDeviceHumifier", a, c)
}
