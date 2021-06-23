package main

import (
	"os"

	"github.com/yomorun/yomo-source-mqtt-starter/pkg/receiver"
)

func main() {
	receiver.CreateRunner("yomo-source", os.Getenv("YOMO_SOURCE_MQTT_ZIPPER_ADDR")).Run()
}
