package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"
	"fmt"

    mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/brutella/hc/log"
)


var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
    log.Info.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
    log.Info.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
    log.Info.Printf("Connect lost: %v", err)
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

	if Config.Debug {
		log.Debug.Enable()
	}

    var broker = "localhost"
    var port = 1883
    opts := mqtt.NewClientOptions()
    opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
    opts.SetDefaultPublishHandler(messagePubHandler)
    opts.OnConnect = connectHandler
    opts.OnConnectionLost = connectLostHandler
    mqtt_cli := mqtt.NewClient(opts)
    if token := mqtt_cli.Connect(); token.Wait() && token.Error() != nil {
        log.Info.Fatal(token.Error())
    }

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	termChan := make(chan int)

	ctx, shutdown := context.WithCancel(context.Background())

	for _, a := range Config.Accessories {
		homekit := makeHomekit(a)
		homekit.Init(mqtt_cli)
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
