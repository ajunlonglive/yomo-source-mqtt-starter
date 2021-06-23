package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/yomorun/yomo-source-mqtt-starter/pkg/receiver"

	"github.com/yomorun/y3-codec-golang"
)

type NoiseData struct {
	Noise float32 `y3:"0x11"` // Noise value
	Time  int64   `y3:"0x12"` // Timestamp (ms)
	From  string  `y3:"0x13"` // Source IP
}

func main() {
	handler := func(topic string, payload []byte, writer receiver.ISourceWriter) error {
		log.Printf("topic=%v, payload=%v\n", topic, string(payload))

		// get data from MQTT
		var raw map[string]int32
		err := json.Unmarshal(payload, &raw)
		if err != nil {
			return err
		}

		// generate y3-codec format
		noise := float32(raw["noise"])
		data := NoiseData{Noise: noise, Time: time.Now().UnixNano() / 1e6, From: "127.0.0.1"}
		sendingBuf, err := y3.NewCodec(0x10).Marshal(data)
		if err != nil {
			return err
		}

		// send data to zipper
		n := 0
		l := len(sendingBuf)
		for n < l {
			n, err = writer.Write(sendingBuf[n:l])
			if err != nil {
				log.Printf("stream.Write error: %v, sendingBuf=%#x\n", err, sendingBuf)
				return err
			}
		}

		log.Printf("write: sendingBuf=%#v\n", sendingBuf)

		return nil
	}

	receiver.CreateRunner("yomo-source", os.Getenv("YOMO_SOURCE_MQTT_ZIPPER_ADDR")).
		WithServerAddr(os.Getenv("YOMO_SOURCE_MQTT_SERVER_ADDR")).
		WithHandler(handler).
		Run()
}
