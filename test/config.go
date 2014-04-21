package main

import (
	"flag"
	"github.com/Terry-Mao/goconf"
	"github.com/golang/glog"
)

var (
	// global config object
	goConf   = goconf.New()
	MyConf   *Config
	confPath string
)

type Config struct {
	RPCAddr  string `goconf:"base:rpc.addr:,"`
	WorkerId int64  `goconf:"base:worker"`
}

func init() {
	flag.StringVar(&confPath, "conf", "./test.conf", " set gosnowflake config file path")
}

// Init init the configuration file.
func InitConfig() error {
	MyConf = &Config{
		RPCAddr:  "localhost:8080",
		WorkerId: int64(0),
	}
	if err := goConf.Parse(confPath); err != nil {
		glog.Errorf("goconf.Parse(\"%s\") error(%v)", confPath, err)
		return err
	}
	if err := goConf.Unmarshal(MyConf); err != nil {
		glog.Errorf("goconf.Unmarshall() error(%v)", err)
		return err
	}
	return nil
}
