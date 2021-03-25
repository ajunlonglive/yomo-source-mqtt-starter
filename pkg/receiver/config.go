package receiver

type Config struct {
	Host   string `json:"host"`
	Port   string `json:"port"`
	Worker int    `json:"workerNum"`
	Debug  bool   `json:"debug"`
}
