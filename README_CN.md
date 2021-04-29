# MQTT-API兼容的YoMo-Source

通过兼容[MQTT](https://mqtt.org/mqtt-specification/)的API协议，YoMo-Source与支持该协议的IoT设备进行连接，并实时高效地把数据以QUIC流的形式传输到YCloud云或者其它部署了YoMo-Zipper的节点。

![schema](https://github.com/yomorun/yomo-source-mqtt-starter/blob/main/docs/schema.jpg?raw=true)

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

- YOMO_SOURCE_MQTT_ZIPPER_ADDR 设置远程yomo-zipper的服务地址。
- YOMO_SOURCE_MQTT_SERVER_ADDR 设置本yomo-source的对外服务地址。
- 发送的数据需要使用y3-codec进行编码后再进行传输，通过定义一个结构体NoiseData传输更多的信息。

#### 3. 创建app.go

```bash
YOMO_SOURCE_MQTT_ZIPPER_ADDR=localhost:9999 YOMO_SOURCE_MQTT_SERVER_ADDR=0.0.0.0:1883 go run app.go
```

## 如何构建镜像

官方提供了一个基础镜像：`yomorun/quic-mqtt:latest`，使用这个镜像可以轻松地部署[YoMo](https://github.com/yomorun/yomo) Source 服务， 以接收来自MQTT协议设备的数据。详细的构建步骤请查看 [yomorun/quic-mqtt](https://hub.docker.com/repository/docker/yomorun/quic-mqtt).



