package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/yomorun/y3-codec-golang"
	"github.com/yomorun/yomo-source-mqtt-starter/pkg/utils"

	"github.com/yomorun/yomo-source-mqtt-starter/pkg/receiver"
	"github.com/yomorun/yomo/pkg/quic"

	"github.com/yomorun/yomo-source-mqtt-starter/internal/env"
)

var (
	zipperAddr          = env.GetString("YOMO_SOURCE_MQTT_ZIPPER_ADDR", "localhost:9999")
	serverAddr          = env.GetString("YOMO_SOURCE_MQTT_SERVER_ADDR", "0.0.0.0:1883")
	serverDebug         = env.GetBool("YOMO_SOURCE_MQTT_SERVER_DEBUG", false)
	streamErrorMax      = env.GetInt("YOMO_SOURCE_MQTT_STREAM_ERROR_MAX", 20)
	streamErrorInterval = env.GetInt("YOMO_SOURCE_MQTT_STREAM_ERROR_INTERVAL", 1000)

	client       quic.Client
	stream       = createStream()
	mutexStream  sync.Mutex
	mutexHandler sync.Mutex
)

type NoiseData struct {
	Noise float32 `y3:"0x11"` // Noise value
	Time  int64   `y3:"0x12"` // Timestamp (ms)
	From  string  `y3:"0x13"` // Source IP
}

func handler(topic string, payload []byte) {
	log.Printf("receive: topic=%v, payload=%v\n", topic, string(payload))

	// get data from MQTT
	var raw map[string]int32
	err := json.Unmarshal(payload, &raw)
	if err != nil {
		log.Printf("Unmarshal payload error:%v", err)
	}

	// generate y3-codec format
	noise := float32(raw["noise"])
	data := NoiseData{Noise: noise, Time: utils.Now(), From: utils.IpAddr()}
	sendingBuf, _ := y3.NewCodec(0x10).Marshal(data)

	mutexHandler.Lock()
	n := 0
	l := len(sendingBuf)
	for n < l {
		n, err = stream.Write(sendingBuf[n:l])
		if err != nil {
			log.Printf("stream.Write error: %v, sendingBuf=%#x\n", err, sendingBuf)
			err = stream.Close()
			if err != nil {
				log.Printf("stream.Close error: %v\n", err)
			}
			stream = createStream()
			break
		}
	}
	mutexHandler.Unlock()

	log.Printf("write: sendingBuf=%#v\n", sendingBuf)
}

func main() {
	receiver.Run(handler, &receiver.Config{ServerAddr: serverAddr, Debug: serverDebug})
}

func createStream() quic.Stream {
	var (
		err    error
		errs   = 0
		stream quic.Stream
	)

	mutexStream.Lock()
	defer mutexStream.Unlock()

CLIENT:
	if client == nil {
		for {
			client, err = quic.NewClient(zipperAddr)
			if err != nil {
				log.Printf("NewClient error: %v, addr=%v\n", err, zipperAddr)
				continue
			}
			break
		}
	}

	for {
		stream, err = client.CreateStream(context.Background())
		if err != nil {
			log.Printf("CreateStream error: %v\n", err)
			errs++
			if errs > streamErrorMax {
				// if greater than the number of errors, a new connection is established
				client = nil
				goto CLIENT
			}
			time.Sleep(time.Duration(streamErrorInterval) * time.Millisecond)
			continue
		}
		break
	}

	return stream
}
