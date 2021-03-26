package receiver

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"time"

	"go.uber.org/zap"
)

func Run(handler func(topic string, payload []byte), config *Config) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if config == nil {
		config = DefaultConfig
	}

	if handler == nil {
		handler = func(topic string, payload []byte) {
			fmt.Printf("%v:\t receive: topic=%v, payload=%v\n", time.Now().Format("2006-01-02 15:04:05"), topic, string(payload))
		}
	}

	re, err := NewReceiver(config)
	if err != nil {
		log.Fatal("New Receiver error", zap.Error(err))
	}
	re.Start(handler)

	s := waitForSignal()
	log.Warn("signal received, receiver closed.", zap.Any("signal", s))
}

func waitForSignal() os.Signal {
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)
	signal.Notify(signalChan, os.Kill, os.Interrupt)
	s := <-signalChan
	signal.Stop(signalChan)
	return s
}
