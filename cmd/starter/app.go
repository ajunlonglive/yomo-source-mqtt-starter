package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/yomorun/yomo-source-mqtt-starter/pkg/receiver"
)

type NoiseData struct {
	Noise float32 `json:"noise"` // Noise value
	Time  int64   `json:"time"` // Timestamp (ms)
	From  string  `json:"from"` // Source IP
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

		// generate json-codec format
		noise := float32(raw["noise"])
		data := NoiseData{Noise: noise, Time: time.Now().UnixNano() / 1e6, From: "127.0.0.1"}
		sendingBuf, err := json.Marshal(data)
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
