package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/brutella/hc"
	"github.com/brutella/hc/log"
)

type ConfigurationDevice struct {
	Name         string
	Shortname    string
	Serialnumber string
	Version      string
	Key          string
	Ip           string
}

type Configuration struct {
	Pin     string
	Devices []ConfigurationDevice
}

func mapInt(x, in_min, in_max, out_min, out_max int) int {
	return (x-in_min)*(out_max-out_min)/(in_max-in_min) + out_min
}

func main() {
	c := flag.String("c", "hk.json", "Specify the configuration file.")
	flag.Parse()
	file, err := os.Open(*c)
	if err != nil {
		log.Info.Fatal("can't open config file: ", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	Config := Configuration{}
	err = decoder.Decode(&Config)
	if err != nil {
		log.Info.Fatal("can't decode config JSON: ", err)
	}
	log.Info.Printf("%v\n", Config)

	hum1 := NewAccessoryHumidifier(&Config.Devices[0])

	t, err := hc.NewIPTransport(hc.Config{Pin: Config.Pin}, hum1.Accessory)
	if err != nil {
		log.Info.Panic(err)
	}

	hc.OnTermination(func() {
		<-t.Stop()
	})

	t.Start()
}
