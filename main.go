package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brutella/hc/log"
)

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

	if Config.Debug {
		log.Debug.Enable()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	termChan := make(chan int)

	ctx, shutdown := context.WithCancel(context.Background())

	for _, a := range Config.Accessories {
		homekit := makeHomekit(a)
		homekit.Init()
		go homekit.Start(ctx)
	}

	for {
		select {
		case <-termChan:
		case s := <-sig:
			log.Info.Printf("Received signal %s, stopping ", s)
			shutdown()

			// TODO wait accessories to stop
			time.Sleep(time.Second * 2)
			return
		}
	}
}
