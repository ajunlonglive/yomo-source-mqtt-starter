package receiver

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"time"

	"go.uber.org/zap"
)

type Runner struct {
	handler func(topic string, payload []byte, writer ISourceWriter) error
	config  *Config

	stream ISourceClient
}

func CreateRunner(appName string, zipperAddr string) *Runner {
	rb := &Runner{
		handler: func(topic string, payload []byte, writer ISourceWriter) error {
			fmt.Printf("%v:\t receive: topic=%v, payload=%v\n", time.Now().Format("2006-01-02 15:04:05"), topic, string(payload))
			fmt.Printf("%v:\t please provide your handler...\n", time.Now().Format("2006-01-02 15:04:05"))
			return nil
		},
		stream: NewSourceStream(appName, zipperAddr),
	}

	rb.config = DefaultConfig

	return rb
}

func (b *Runner) WithHandler(handler func(topic string, payload []byte, writer ISourceWriter) error) *Runner {
	b.handler = handler
	return b
}

func (b *Runner) WithServerAddr(addr string) *Runner {
	b.config.ServerAddr = addr
	return b
}

func (b *Runner) WithDebug(debug bool) *Runner {
	b.config.Debug = debug
	return b
}

func (b *Runner) WithStream(stream ISourceClient) *Runner {
	b.stream = stream
	return b
}

func (b *Runner) Run() {
	b.runWithBlock()
}

func (b *Runner) runWithBlock() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	re, err := NewReceiver(b.config)
	if err != nil {
		log.Fatal("New Receiver error", zap.Error(err))
	}

	b.stream.Init() //预连接

	re.Start(b.handler, b.stream)

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
