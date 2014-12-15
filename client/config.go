package client

import (
	"flag"
	"github.com/Terry-Mao/goconf"
	"time"
)

var (
	// global config object
	goConf   = goconf.New()
	MyConf   *Config
	confPath string
)

type Config struct {
	RPCAddr   string        `goconf:"base:rpc.addr:,"`
	WorkerId  int64         `goconf:"base:worker"`
	ZKServers []string      `goconf:"zookeeper:addr:,"`
	ZKPath    string        `goconf:"zookeeper:path"`
	ZKTimeout time.Duration `goconf:"zookeeper:timeout:time"`
}

func init() {
	flag.StringVar(&confPath, "conf", "./test.conf", " set gosnowflake config file path")
}

// Init init the configuration file.
func InitConfig() error {
	MyConf = &Config{
		RPCAddr:   "localhost:8080",
		WorkerId:  int64(0),
		ZKServers: []string{"localhost:2181"},
		ZKPath:    "/gosnowflake-servers",
		ZKTimeout: time.Second * 15,
	}
	if err := goConf.Parse(confPath); err != nil {
		return err
	}
	if err := goConf.Unmarshal(MyConf); err != nil {
		return err
	}
	return nil
}
