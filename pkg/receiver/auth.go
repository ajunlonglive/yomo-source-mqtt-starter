package receiver

import "github.com/yomorun/yomo-source-mqtt-starter/internal/env"

var (
	enableAuth     = env.GetBool("CONNECTOR_MQTT_AUTH_ENABLE", false)
	serverUserName = env.GetString("CONNECTOR_MQTT_AUTH_USERNAME", "admin")
	serverPassword = env.GetString("CONNECTOR_MQTT_AUTH_PASSWORD", "public")
)

func (b *Receiver) CheckConnectAuth(clientID, username, password string) bool {
	if enableAuth {
		return serverUserName == username && serverPassword == password
	}
	return true
}
