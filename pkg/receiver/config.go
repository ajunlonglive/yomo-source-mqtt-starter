package receiver

var (
	DefaultConfig *Config
)

func init() {
	DefaultConfig = &Config{ServerAddr: "0.0.0.0:1883", Worker: 4096, Debug: false}
}

type Config struct {
	ServerAddr string `json:"serverAddr"`
	Worker     int    `json:"workerNum"`
	Debug      bool   `json:"debug"`
}
