package main

import (
	"flag"
	"github.com/Terry-Mao/goconf"
	"github.com/golang/glog"
	"runtime"
	"time"
)

var (
	// global config object
	goConf   = goconf.New()
	MyConf   *Config
	confPath string
)

type Config struct {
	User         string        `goconf:"base:user"`
	PidFile      string        `goconf:"base:pid"`
	Dir          string        `goconf:"base:dir"`
	MaxProc      int           `goconf:"base:maxproc"`
	RPCBind      []string      `goconf:"base:rpc.bind:,"`
	StatBind     []string      `goconf:"base:stat.bind:,"`
	PprofBind    []string      `goconf:"base:pprof.bind:,"`
	DatacenterId int64         `goconf:"snowflake:datacenter"`
	WorkerId     []int64       `goconf:"snowflake:worker"`
	ZKAddr       []string      `goconf:"zookeeper:addr"`
	ZKTimeout    time.Duration `goconf:"zookeeper:timeout:time`
	ZKPath       string        `goconf:"zookeeper:path"`
}

func init() {
	flag.StringVar(&confPath, "conf", "./gosnowflake.conf", " set gosnowflake config file path")
}

// Init init the configuration file.
func InitConfig() error {
	MyConf = &Config{
		User:         "nobody",
		PidFile:      "/tmp/user_account.pid",
		Dir:          "/dev/null",
		MaxProc:      runtime.NumCPU(),
		RPCBind:      []string{"localhost:8080"},
		DatacenterId: 0,
		WorkerId:     []int64{0},
		ZKAddr:       []string{"localhost:2181"},
		ZKTimeout:    time.Second * 15,
		ZKPath:       "/gosnowflake-servers",
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

// Reload reload the configuration file.
func ReloadConfig() error {
	glog.Infof("config file: \"%s\" reload", confPath)
	t, err := goConf.Reload()
	if err != nil {
		glog.Errorf("goconf.Reload() error(%v)", err)
		glog.Warningf("confi file: \"%s\" reload failed, use original one", confPath)
		return err
	}
	goConf = t
	// new a Config for Unmarshal
	myConf := &Config{}
	if err := goConf.Unmarshal(myConf); err != nil {
		glog.Errorf("goconf.Unmarshall() error(%v)", err)
		return err
	}
	// atomic replace MyConf
	MyConf = myConf
	return nil
}
