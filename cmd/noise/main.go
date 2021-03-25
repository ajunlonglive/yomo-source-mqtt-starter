package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"

	"github.com/yomorun/yomo-source-mqtt-starter/pkg/receiver"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	config := &receiver.Config{Host: "127.0.0.1", Port: "8888", Worker: 4096, Debug: true}

	b, err := receiver.NewReceiver(config)
	if err != nil {
		log.Fatal("New Receiver error: ", err)
	}
	b.Start(func(topic string, payload []byte) {
		log.Printf("receive: topic=%v, payload=%v\n", topic, string(payload))
	})

	s := waitForSignal()
	log.Println("signal received, receiver closed.", s)
}

func waitForSignal() os.Signal {
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)
	signal.Notify(signalChan, os.Kill, os.Interrupt)
	s := <-signalChan
	signal.Stop(signalChan)
	return s
}
