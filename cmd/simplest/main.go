package main

import (
	"github.com/yomorun/yomo-source-mqtt-starter/pkg/receiver"
)

func main() {
	receiver.CreateRunner("localhost:9999").Run()
}
