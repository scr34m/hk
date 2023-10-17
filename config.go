package main

type ConfigurationDevice struct {
	Name         string
	Manufacturer string
	Serialnumber string
	Version      string
	Key          string
	Ip           string
	Url          string
	Monitor_18   bool
	Monitor_19   bool
}

type ConfigurationAccessory struct {
	Pin          string
	Type         string
	Internalname string
	Device       ConfigurationDevice
}

type Configuration struct {
	Debug       bool
	Accessories []ConfigurationAccessory
}
