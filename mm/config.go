package mm

import (
	"os"
	"runtime"

	"github.com/pquerna/ffjson/ffjson"
)

//go:generate ffjson $GOFILE

// ffjson: noencoder
type Configuration struct {
	Ip       string    `json:"server.ip"`
	Port     uint32    `json:"server.port"`
	DbPath   *Pathname `json:"database.file"`
	Threads  uint8     `json:"worker.threads"`
	CacheTtl float64   `json:"cache.ttl"`
}

// create a new configuration with default values
func NewConfiguration() *Configuration {
	return &Configuration{
		Ip:       "127.0.0.1",
		Port:     8000,
		DbPath:   NewPathname("/var/lib/maxminddb/GeoLite2-City.mmdb"),
		Threads:  uint8(runtime.NumCPU()),
		CacheTtl: float64(3600),
	}
}

func (c *Configuration) LoadFromJson(data []byte) error {
	if err := ffjson.Unmarshal(data, c); err != nil {
		return err
	}
	return nil
}

func (c *Configuration) LoadFromJsonFile(configFile *Pathname) error {
	if !configFile.Exists() {
		configFile = NewPathname(os.Args[0]).Dirname().Join("config.json")
		if !configFile.Exists() {
			return nil
		}
	}

	b, err := configFile.Read()
	if err != nil {
		return err
	}

	return c.LoadFromJson(b)
}
