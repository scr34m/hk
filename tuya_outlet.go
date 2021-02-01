package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/scr34m/tuya"
)

type TuyaDeviceOutlet interface {
	Set(string, interface{}) error
	SetW(string, interface{}, time.Duration) (interface{}, error)
	Status(string) interface{}
}

type ITuyaDeviceOutlet struct {
	tuya.BaseDevice
	// 1	Switch 1	bool	True/False
	// 9	Countdown 1	integer	0-86400	s
	// 17	Add Electricity	integer	0-50000	kwh
	// 18	Current	integer	0-30000	mA
	// 19	Power	integer	0-50000	W
	// 20	Voltage	integer	0-5000	V
	// 21	Test Bit	integer	0-5	n/a
	// 22	Voltage coe	integer	0-1000000
	// 23	Current coe	integer	0-1000000
	// 24	Power coe	integer	0-1000000
	// 25	Electricity coe	integer	0-1000000
	// 26	Fault	fault	ov_cr
	status map[string]interface{}
}

func (s *ITuyaDeviceOutlet) Set(dps string, value interface{}) error {
	m := s.App.MakeBaseMsg()
	m["dps"] = map[string]interface{}{dps: value}
	return s.App.SendEncryptedCommand(tuya.CodeMsgSet, m)
}

func (s *ITuyaDeviceOutlet) SetW(dps string, value interface{}, delay time.Duration) (interface{}, error) {
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

func (s *ITuyaDeviceOutlet) Status(key string) interface{} {
	if val, ok := s.status[key]; ok {
		return val
	} else {
		return nil
	}
}

func (s *ITuyaDeviceOutlet) ProcessResponse(code int, data []byte) {
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
	log.Printf("%v\n", r)
	s.Notify(code, s)
}

func (s *ITuyaDeviceOutlet) Configure(a *tuya.Appliance, c *tuya.ConfigurationData) {
	s.status = make(map[string]interface{})
	s.Init("TuyaDeviceOutlet", a, c)
}
