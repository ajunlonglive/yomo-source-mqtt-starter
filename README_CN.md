# MQTT-API兼容的YoMo-Source

通过兼容[MQTT](https://mqtt.org/mqtt-specification/)的API协议，YoMo-Source与支持该协议的IoT设备进行连接，并实时高效地把数据以QUIC流的形式传输到YCloud云或者其它部署了YoMo-Zipper的节点。

![schema](./docs/schema.jpg)

## 🚀 快速入门

### 例子 (噪音)

在这个例子中，假设`噪音传感器`以MQTT协议的方式向外传输主题为`NOISE`的器音数据，格式如下：

```json
{"noise":416}
```

那么，我们可以通过引用[yomo-source-mqtt-starter](https://github.com/yomorun/yomo-source-mqtt-starter)组件来创建一个yomo-source来接收噪音传感器发送的数据，并传输给部署了yomo-zipper服务的云端。

#### 1. 初始化项目

```bash
go mod init source
go get github.com/yomorun/yomo-source-mqtt-starter
```

#### 2. 创建app.go

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yomorun/yomo-source-mqtt-starter/pkg/receiver"

	"github.com/yomorun/yomo/pkg/quic"

	"github.com/yomorun/y3-codec-golang"
)

type NoiseData struct {
	Noise float32 `y3:"0x11"` // Noise value
	Time  int64   `y3:"0x12"` // Timestamp (ms)
	From  string  `y3:"0x13"` // Source IP
}

func main() {
	client, err := quic.NewClient(os.Getenv("YOMO_SOURCE_MQTT_ZIPPER_ADDR"))
	if err != nil {
		panic(fmt.Errorf("NewClient error: %v", err))
	}

	stream, err := client.CreateStream(context.Background())
	if err != nil {
		panic(fmt.Errorf("CreateStream error:%s", err.Error()))
	}

	handler := func(topic string, payload []byte) {
		log.Printf("topic=%v, payload=%v\n", topic, string(payload))

		// get data from MQTT
		var raw map[string]int32
		err := json.Unmarshal(payload, &raw)
		if err != nil {
			log.Printf("Unmarshal payload error:%v", err)
		}

		// generate y3-codec format
		noise := float32(raw["noise"])
		data := NoiseData{Noise: noise, Time: time.Now().UnixNano() / 1e6, From: "127.0.0.1"}
		sendingBuf, _ := y3.NewCodec(0x10).Marshal(data)

		// send data to zipper
		n := 0
		l := len(sendingBuf)
		for n < l {
			n, err = stream.Write(sendingBuf[n:l])
			if err != nil {
				log.Printf("stream.Write error: %v, sendingBuf=%#x\n", err, sendingBuf)
			}
		}
	}

	receiver.Run(handler, &receiver.Config{ServerAddr: os.Getenv("YOMO_SOURCE_MQTT_SERVER_ADDR")})
}
```

- YOMO_SOURCE_MQTT_ZIPPER_ADDR 设置远程yomo-zipper的服务地址。
- YOMO_SOURCE_MQTT_SERVER_ADDR 设置本yomo-source的对外服务地址。
- 发送的数据需要使用y3-codec进行编码后再进行传输，通过定义一个结构体NoiseData传输更多的信息。

#### 3. 创建app.go

```bash
YOMO_SOURCE_MQTT_ZIPPER_ADDR=localhost:9999 YOMO_SOURCE_MQTT_SERVER_ADDR=0.0.0.0:1883 go run app.go
```

