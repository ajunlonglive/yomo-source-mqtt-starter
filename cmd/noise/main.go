package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/yomorun/y3-codec-golang"
	"github.com/yomorun/yomo-source-mqtt-starter/pkg/utils"

	"github.com/yomorun/yomo-source-mqtt-starter/pkg/receiver"
)

type NoiseData struct {
	Noise float32 `y3:"0x11"` // Noise value
	Time  int64   `y3:"0x12"` // Timestamp (ms)
	From  string  `y3:"0x13"` // Source IP
}

func main() {
	handler := func(topic string, payload []byte, writer receiver.ISourceWriter) error {
		log.Printf("receive: topic=%v, payload=%v\n", topic, string(payload))

		// 1.get data from MQTT
		var raw map[string]int32
		err := json.Unmarshal(payload, &raw)
		if err != nil {
			log.Printf("Unmarshal payload error:%v", err)
		}

		// 2.generate y3-codec format
		noise := float32(raw["noise"])
		data := NoiseData{Noise: noise, Time: utils.Now(), From: utils.IpAddr()}
		sendingBuf, _ := y3.NewCodec(0x10).Marshal(data)

		// 3.send data to remote workflow engine
		_, err = writer.Write(sendingBuf)
		if err != nil {
			log.Printf("stream.Write error: %v, sendingBuf=%#x\n", err, sendingBuf)
			return err
		}

		log.Printf("write: sendingBuf=%#v\n", sendingBuf)
		return nil
	}

	receiver.CreateRunner(os.Getenv("YOMO_SOURCE_MQTT_ZIPPER_ADDR")).
		WithServerAddr(os.Getenv("YOMO_SOURCE_MQTT_SERVER_ADDR")).
		WithHandler(handler).
		Run()
}
