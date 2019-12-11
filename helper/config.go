package helper

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

const (
	PIPA_CONF_PATH = "/etc/yig/pipa/pipa.toml"
	PIAP_FRONT_PATh = "/usr/share/fonts/Chinese_fonts/"
)

type Config struct {
	LogLevel       string `toml:"log_level"`
	LogPath        string `toml:"log_path"`
	BindApiAddress string `toml:"api_listener"`

	RedisAddress         string `toml:"redis_address"`  // redis connection string, e.g localhost:1234
	RedisPassword        string `toml:"redis_password"` // redis auth password
	RedisConnectTimeout  int    `toml:"redis_connect_timeout"`
	RedisReadTimeout     int    `toml:"redis_read_timeout"`
	RedisWriteTimeout    int    `toml:"redis_write_timeout"`
	RedisPoolMaxIdle     int    `toml:"redis_pool_max_idle"`
	RedisPoolIdleTimeout int    `toml:"redis_pool_idle_timeout"`

	FactoryWorkersNumber int	`toml:"factory_workers_number"`
}

var CONFIG Config

func SetupConfig() {
	MarshalTOMLConfig()
}

func MarshalTOMLConfig() error {
	data, err := ioutil.ReadFile(PIPA_CONF_PATH)
	if err != nil {
		if err != nil {
			panic("Cannot open yig.toml")
		}
	}
	var c Config
	_, err = toml.Decode(string(data), &c)
	if err != nil {
		panic("load yig.toml error: " + err.Error())
	}

	return nil
}
