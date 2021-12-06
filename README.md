# mqtt-api-compatible yomo-source
[MQTT](https://mqtt.org/mqtt-specification/) protocol-enabled IoT devices connect to YoMo-Source and efficiently transfer data in real-time as QUIC streams to the YCloud cloud or other nodes where YoMo-Zipper is deployed.

![schema](https://github.com/yomorun/yomo-source-mqtt-starter/blob/main/docs/schema.jpg?raw=true)

## 🚀 Getting Started

### Example (noise)

This example shows how to use the component reference method to make it easier to receive MQTT messages using starter and convert them to the YoMo protocol for transmission to the Zipper service.

#### 1. Init Project

```bash
go mod init source
go get github.com/yomorun/yomo-source-mqtt-starter
```

#### 2. create app.go

```go
package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/yomorun/yomo-source-mqtt-starter/pkg/utils"
	"github.com/yomorun/yomo-source-mqtt-starter/pkg/receiver"
)

type NoiseData struct {
	Noise float32 `json:"noise"` // Noise value
	Time  int64   `json:"time"` // Timestamp (ms)
	From  string  `json:"from"` // Source IP
}

func main() {
	handler := func(topic string, payload []byte, writer receiver.ISourceWriter) error {
		log.Printf("receive: topic=%v, payload=%v\n", topic, string(payload))

		// get data from MQTT
		var raw map[string]int32
		err := json.Unmarshal(payload, &raw)
		if err != nil {
			log.Printf("Unmarshal payload error:%v", err)
		}

		// generate json-codec format
		noise := float32(raw["noise"])
		data := NoiseData{Noise: noise, Time: utils.Now(), From: utils.IpAddr()}
		sendingBuf, _ := json.Marshal(data)

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
```

- YOMO_SOURCE_MQTT_ZIPPER_ADDR: Set the service address of the remote yomo-zipper.
- YOMO_SOURCE_MQTT_SERVER_ADDR: Set the external service address of this yomo-source.
- The data to be sent needs to be encoded using y3-codec.

#### 3. run

```go
YOMO_SOURCE_MQTT_ZIPPER_ADDR=localhost:9999 YOMO_SOURCE_MQTT_SERVER_ADDR=0.0.0.0:1883 go run app.go
```

## How to build the Image

Official base image is `yomorun/quic-mqtt:latest`, using this image you can easily deploy the [YoMo](https://github.com/yomorun/yomo) Source service for receiving data from MQTT protocol devices. For detailed build steps, please see [yomorun/quic-mqtt](https://hub.docker.com/repository/docker/yomorun/quic-mqtt).

