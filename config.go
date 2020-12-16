package main

type ConfigurationDevice struct {
	Name         string
	Manufacturer string
	Serialnumber string
	Version      string
	Key          string
	Ip           string
	Url          string
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
