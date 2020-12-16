package main

type ConfigurationDevice struct {
	Name         string
	Shortname    string
	Serialnumber string
	Version      string
	Key          string
	Ip           string
}

type ConfigurationAccessory struct {
	Pin    string
	Type   string
	Device ConfigurationDevice
}

type Configuration struct {
	Debug       bool
	Accessories []ConfigurationAccessory
}
